# Multi-Turn Agent Architecture

This document describes how the Agentic RAG Pokédex processes each trainer message: session lifecycle, tool-calling loop, memory model, and UI trace events.

Entry point: `AskPokedex()` in `app.go` → `agent.Loop.Run()`.

## High-Level Architecture

```mermaid
flowchart TB
    subgraph UI["Wails Frontend"]
        Composer["Trainer input<br/>(text + optional image)"]
        Feed["Agent Feed<br/>(trace timeline)"]
        Answer["Answer stream"]
    end

    subgraph App["app.go"]
        Ask["AskPokedex()"]
    end

    subgraph Loop["agent.Loop"]
        Run["Run()"]
        Lifecycle["runSessionLifecycle()"]
        ToolLoop["Tool-calling loop<br/>(max N iterations)"]
        Fallback["runPrefetchFallback()"]
    end

    subgraph Session["session.Manager"]
        Ensure["EnsureActiveSession"]
        Decide["Decider.Decide()"]
        Apply["ApplyDecision()"]
        Battle["UpdateBattleState /<br/>TryBattleRecommendation"]
        Context["SessionContextPrompt()"]
        Save["SaveTurn()"]
    end

    subgraph LLM["llm.Client (Ollama)"]
        Chat["GenerateWithTools()"]
        Stream["StreamChat()"]
        Vision["DescribeImage()"]
        Complete["Complete() — session decisions"]
    end

    subgraph Tools["agent.Registry"]
        Clock["clock"]
        GPS["gps"]
        PokeDB["pokemon_db"]
        Mem["session_memory"]
        Obs["record_observation"]
        RAG["knowledge_search"]
    end

    subgraph Stores["Persistence"]
        SessionDB["Session Store<br/>(SQLite + embeddings)"]
        TeamDB["Team Store<br/>(SQLite)"]
        VectorDB["RAG Store<br/>(Kanto corpus)"]
        Sim["Simulation<br/>(Clock + GPS)"]
    end

    Composer --> Ask
    Ask --> Run
    Run --> Ensure
    Ensure --> Lifecycle
    Lifecycle --> Vision
    Lifecycle --> Decide
    Decide --> Complete
    Lifecycle --> Apply
    Apply --> SessionDB
    Apply --> TeamDB
    Lifecycle --> Battle
    Battle --> SessionDB
    Lifecycle -->|session context| Context
    Context --> SessionDB

    Lifecycle -->|early exit| Save
    Lifecycle -->|continue| ToolLoop

    ToolLoop --> Chat
    Chat -->|tool calls| Tools
    Tools --> Clock & GPS & PokeDB & Mem & Obs & RAG
    Clock & GPS --> Sim
    PokeDB --> TeamDB
    Mem & Obs --> SessionDB
    RAG --> VectorDB

    Chat -->|no tools| Stream
    Chat -->|tools unsupported| Fallback
    Fallback --> Tools

    ToolLoop --> Save
    Stream --> Answer
    Run -->|onTrace| Feed
    Save --> SessionDB
```

## Single-Turn Flow

Each trainer message passes through up to three phases. Several paths short-circuit before the main tool loop.

```mermaid
flowchart TD
    Start(["Trainer message"]) --> Input["Decode text / image"]
    Input --> Session["EnsureActiveSession()"]

    Session --> Phase1{"Phase 1:<br/>Session lifecycle"}

    Phase1 --> Img{"Has image?"}
    Img -->|yes| Vision["LLM: DescribeImage()"]
    Img -->|no| BuildCtx
    Vision --> BuildCtx["BuildDecisionContext()<br/>recent events, team, GPS, battle state"]

    BuildCtx --> Decide["Decider.Decide()"]
    Decide --> Heur1["Health heuristics"]
    Heur1 -->|match| Decision
    Heur1 -->|no match| LLMDecide["LLM → JSON decision"]
    LLMDecide --> Heur2["Health override check"]
    Heur2 --> Decision["AgentSessionDecision"]
    LLMDecide -->|fail| Heuristic["Heuristic fallback"]
    Heuristic --> Decision

    Decision --> Apply["ApplyDecision()<br/>log events, update HP,<br/>close/compact/start session"]

    Apply --> Clarify{"Needs<br/>clarification?"}
    Clarify -->|yes| BattleRec1{"Battle<br/>recommendation?"}
    BattleRec1 -->|yes| Out1["Stream answer → SaveTurn → END"]
    BattleRec1 -->|no| Out2["Stream clarification → SaveTurn → END"]

    Clarify -->|no| BattleState["UpdateBattleState()"]
    BattleState --> BattleRec2{"Battle<br/>recommendation?"}
    BattleRec2 -->|yes| Out3["Stream answer → SaveTurn → END"]
    BattleRec2 -->|no| Ctx["SessionContextPrompt()<br/>inject into system prompt"]

    Ctx --> Phase2{"Phase 2:<br/>Tool-calling loop<br/>(≤ max iterations)"}

    Phase2 --> Gen["LLM GenerateWithTools()"]
    Gen -->|error: no tool support| Prefetch["Prefetch all relevant tools<br/>then StreamChat()"]
    Prefetch --> Out4["Stream answer → SaveTurn → END"]

    Gen --> HasTools{"Tool calls<br/>in response?"}
    HasTools -->|no| Final["Stream final answer"]
    Final --> Out5["SaveTurn → END"]

    HasTools -->|yes| Plan["planToolCall()<br/>route corrections"]
    Plan --> Exec["Registry.Execute()"]
    Exec --> Append["Append assistant turn +<br/>tool results to messages"]
    Append --> Limit{"Iteration<br/>limit?"}
    Limit -->|no| Gen
    Limit -->|yes| Out6["Fallback message → SaveTurn → END"]
```

## Session Decision Block

Before the agent answers, a separate LLM call (with heuristic fallbacks) decides how to manage session state.

```mermaid
flowchart LR
    subgraph Input
        Msg["User message"]
        ImgDesc["Image description"]
        Events["Recent events (20)"]
        Team["Current party"]
        Loc["GPS location"]
        Battle["Battle state"]
    end

    subgraph Actions["Possible actions"]
        A1["continue_session"]
        A2["add_observation"]
        A3["update_pokemon_state"]
        A4["ask_clarification"]
        A5["close_session + compact"]
        A6["start_new_session"]
    end

    Input --> Decider
    Decider --> Actions

    A3 --> HP["Update team HP in SQLite"]
    A2 --> Obs["Append session event"]
    A5 --> Close["Summarize & end session"]
    A6 --> New["Start fresh session"]
    A4 --> Early["Skip tool loop;<br/>ask trainer or recommend"]
```

### Decision actions

| Action | Effect |
|--------|--------|
| `continue_session` | Keep the active session; no structural change |
| `add_observation` | Record a gameplay event (battle, location, note) |
| `update_pokemon_state` | Update party HP in the team store |
| `ask_clarification` | Return early with a clarification prompt |
| `close_session` / `compact_session` | Summarize and end the current session |
| `start_new_session` | Begin a fresh session (often after a location shift) |

Decision order in `Decider.Decide()`:

1. Health-update heuristics (regex on fainted/injured language)
2. LLM JSON decision via `Complete()`
3. Health override check (heuristics win over LLM for HP updates)
4. Heuristic fallback if the LLM call fails

## Tool-Calling Loop

The core ReAct-style loop runs until the model returns a final answer or hits the iteration cap (`AGENT_MAX_ITERATIONS`, default 5).

| Step | What happens |
|------|----------------|
| **System prompt** | Kanto Pokédex persona + active session context + battle/health hints |
| **Human message** | Trainer text ± image binary part |
| **LLM turn** | `GenerateWithTools()` returns thought text + optional tool calls |
| **Route correction** | e.g. `session_memory` → `pokemon_db` when the question is about the current team |
| **Tool execution** | Registry dispatches to registered tools |
| **Message growth** | Assistant turn + tool results appended → next iteration |
| **Termination** | No tool calls → stream answer; or iteration limit → timeout message |

### Registered tools

| Tool | Purpose | Backing store |
|------|---------|---------------|
| `clock` | In-game time and weather | `simulation.Clock` |
| `gps` | Trainer location | `simulation.GPS` |
| `pokemon_db` | Current party (filter, sort, limit) | `pokemonstore.SQLiteStore` |
| `session_memory` | Semantic search over past sessions | `session.Store` (embeddings) |
| `record_observation` | Save a gameplay event for long-term memory | `session.Store` |
| `knowledge_search` | Pokédex facts and Kanto lore | `rag.Retriever` |

### Prefetch fallback

When the chat model does not support native tool calling (`llm.IsToolsNotSupported`), `runPrefetchFallback()` gathers context automatically (clock, gps, party, memory, knowledge) and streams a plain chat completion instead.

## Trace Events

Each step emits a `TraceStep` to the Agent Feed UI via Wails `agent:trace` events.

| Kind | Typical title | When |
|------|---------------|------|
| `Event` | New Observation, GPS Update, Session Decision | Lifecycle and environment updates |
| `Thought` | Analyze Situation | LLM reasoning text before tool calls |
| `Action` | Use Tools | Model requests one or more tools |
| `Observation` | Retrieval Results | Tool output returned to the model |
| `FinalAnswer` | Response Ready, Battle Recommendation | Answer streamed to the trainer |

When a session reset occurs (close + start new), the trace timeline resets via `agent:trace-reset`.

## Multi-Turn Memory Model

Turns accumulate across the conversation through layered memory.

```mermaid
flowchart TB
    subgraph PerTurn["Each turn"]
        T1["User message saved as event"]
        T2["Observations / state updates saved"]
        T3["Final answer + trace saved via SaveTurn()"]
    end

    subgraph CrossTurn["Cross-turn context"]
        Active["Active session summary"]
        Recent["Last 8 session events"]
        BattleS["Structured battle state<br/>(opponent, their Pokémon)"]
        LastSum["Last ended session summary"]
    end

    subgraph Retrieval["On-demand retrieval (tools)"]
        SM["session_memory — semantic search<br/>over past sessions"]
        PDB["pokemon_db — current party"]
        KS["knowledge_search — Kanto corpus"]
    end

    PerTurn --> CrossTurn
    CrossTurn -->|injected in system prompt| NextTurn["Next turn's Phase 1 & 2"]
    NextTurn --> Retrieval
```

**Injected context** (system prompt, every turn): active session summary, recent events, structured battle state, fainted-Pokémon advisories.

**Retrieved on demand** (tools): past session search, party queries, Kanto lore.

## Key Source Files

| Component | File |
|-----------|------|
| Entry point | `app.go` → `AskPokedex()` |
| Main orchestrator | `internal/agent/loop.go` → `Run()` |
| Session lifecycle | `internal/agent/loop.go` → `runSessionLifecycle()` |
| Session decisions | `internal/session/decision.go` |
| Session apply logic | `internal/session/manager.go` → `ApplyDecision()` |
| Battle state & recommendations | `internal/session/battle_state.go` |
| Tool registry wiring | `internal/agent/wiring.go`, `tool.go` |
| No-tool fallback | `internal/agent/fallback.go` |
| Trace step types | `internal/agent/trace.go` |
| LLM client | `internal/llm/toolchat.go`, `stream.go` |

## Design Summary

The agent separates three concerns:

1. **Autonomous session management** (Phase 1) — runs before every answer; decides observations, HP updates, session boundaries, and clarification without asking the trainer to manage sessions manually.
2. **ReAct tool calling** (Phase 2) — iteratively gathers context via tools until the model can answer.
3. **Battle shortcuts** — when structured battle state is sufficient, `TryBattleRecommendation()` can answer type-matchup advice without entering the full tool loop.
