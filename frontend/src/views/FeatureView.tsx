import {useEffect, useState} from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import type {MenuId} from '../config/menuItems';
import {menuItems} from '../config/menuItems';
import {
    AskPokedex,
    GetCurrentSituation,
    GetKnowledgeBase,
    GetMyTeam,
    GetPokemonList,
    GetTypeAnalysis,
} from '../../wailsjs/go/main/App';
import {EventsOn} from '../../wailsjs/runtime/runtime';
import {domain} from '../../wailsjs/go/models';
import {PokemonDetailPanel} from '../components/PokemonDetailPanel';
import {TeamPokemonCards} from '../components/TeamPokemonCards';

interface FeatureViewProps {
    screen: MenuId;
    onBack: () => void;
}

export function FeatureView({screen, onBack}: FeatureViewProps) {
    const title = menuItems.find((m) => m.id === screen)?.title ?? 'Feature';
    const [content, setContent] = useState<string>('');
    const [busy, setBusy] = useState(false);
    const [askInput, setAskInput] = useState('');
    const [searchQuery, setSearchQuery] = useState('');
    const [analysisTarget, setAnalysisTarget] = useState('Squirtle');
    const [team, setTeam] = useState<domain.Pokemon[]>([]);
    const [selectedPokemon, setSelectedPokemon] = useState<string | null>(null);

    useEffect(() => {
        if (screen !== 'ask-pokedex') return;
        const unsubscribe = EventsOn('llm:chunk', (...args: unknown[]) => {
            const chunk =
                typeof args[0] === 'string' ? args[0] : String(args[0] ?? '');
            setContent((prev) => prev + chunk);
        });
        return unsubscribe;
    }, [screen]);

    useEffect(() => {
        if (screen === 'ask-pokedex' || screen === 'analysis') return;
        loadScreen(screen, searchQuery, analysisTarget);
    }, [screen, searchQuery]);

    useEffect(() => {
        if (screen === 'analysis') {
            loadScreen('analysis', searchQuery, analysisTarget);
        }
    }, [screen]);

    useEffect(() => {
        function handleKeyDown(e: KeyboardEvent) {
            if (e.key !== 'b' && e.key !== 'B') return;
            const target = e.target;
            if (
                target instanceof HTMLInputElement ||
                target instanceof HTMLTextAreaElement ||
                target instanceof HTMLSelectElement ||
                (target instanceof HTMLElement && target.isContentEditable)
            ) {
                return;
            }
            e.preventDefault();
            if (screen === 'my-team' && selectedPokemon) {
                setSelectedPokemon(null);
            } else {
                onBack();
            }
        }
        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, [onBack, screen, selectedPokemon]);

    async function loadScreen(
        id: MenuId,
        query: string,
        pokemonName: string,
    ) {
        setBusy(true);
        setContent('');
        setTeam([]);
        setSelectedPokemon(null);
        try {
            switch (id) {
                case 'current-situation': {
                    const situation = await GetCurrentSituation();
                    setContent(
                        `**${situation.summary}**\n\n${situation.advice.map((a) => `- ${a}`).join('\n')}`,
                    );
                    break;
                }
                case 'my-team': {
                    setTeam(await GetMyTeam());
                    break;
                }
                case 'pokemon': {
                    const list = await GetPokemonList(query);
                    setContent(
                        list
                            .map(
                                (p) =>
                                    `**#${p.id} ${p.name}** — ${p.types} (${p.region})`,
                            )
                            .join('\n\n') || '_No matches._',
                    );
                    break;
                }
                case 'analysis': {
                    const analysis = await GetTypeAnalysis(pokemonName);
                    setContent(formatAnalysis(analysis));
                    break;
                }
                case 'knowledge-base': {
                    const articles = await GetKnowledgeBase();
                    setContent(
                        articles
                            .map((a) => `### ${a.title}\n${a.summary}`)
                            .join('\n\n'),
                    );
                    break;
                }
            }
        } catch {
            setContent('Something went wrong. Try again.');
        } finally {
            setBusy(false);
        }
    }

    async function submitAsk() {
        const prompt = askInput.trim();
        if (!prompt) return;
        setBusy(true);
        setContent('');
        try {
            await AskPokedex(prompt);
        } catch {
            setContent('Could not reach the Pokédex. Try again.');
        } finally {
            setBusy(false);
        }
    }

    const waitingForStream =
        screen === 'ask-pokedex' && busy && content === '';

    return (
        <div className="feature-view">
            <header className="feature-view__header">
                <h1>{title}</h1>
            </header>

            {screen === 'pokemon' && (
                <div className="feature-view__controls">
                    <input
                        type="search"
                        placeholder="Search Pokémon…"
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                    />
                </div>
            )}

            {screen === 'analysis' && (
                <div className="feature-view__controls">
                    <input
                        type="text"
                        placeholder="Pokémon name"
                        value={analysisTarget}
                        onChange={(e) => setAnalysisTarget(e.target.value)}
                    />
                    <button
                        type="button"
                        onClick={() =>
                            loadScreen('analysis', searchQuery, analysisTarget)
                        }
                    >
                        Analyze
                    </button>
                </div>
            )}

            {screen === 'ask-pokedex' && (
                <div className="feature-view__controls feature-view__controls--ask">
                    <input
                        type="text"
                        placeholder="Ask anything…"
                        value={askInput}
                        onChange={(e) => setAskInput(e.target.value)}
                        onKeyDown={(e) => e.key === 'Enter' && submitAsk()}
                    />
                    <button
                        type="button"
                        onClick={submitAsk}
                        disabled={busy || !askInput.trim()}
                    >
                        Ask
                    </button>
                </div>
            )}

            <div className="feature-view__content" role="status">
                {busy && screen === 'my-team' ? (
                    <p className="feature-view__loading">Loading…</p>
                ) : screen === 'my-team' ? (
                    <>
                        <TeamPokemonCards
                            team={team}
                            onSelect={setSelectedPokemon}
                        />
                        {selectedPokemon && (
                            <PokemonDetailPanel
                                name={selectedPokemon}
                                onClose={() => setSelectedPokemon(null)}
                            />
                        )}
                    </>
                ) : busy && screen !== 'ask-pokedex' && !content ? (
                    <p className="feature-view__loading">Loading…</p>
                ) : waitingForStream ? (
                    <p className="feature-view__loading">Thinking…</p>
                ) : (
                    <div className="feature-view__md">
                        <ReactMarkdown remarkPlugins={[remarkGfm]}>
                            {content}
                        </ReactMarkdown>
                        {screen === 'ask-pokedex' && busy && (
                            <span className="poke-stream-caret" aria-hidden />
                        )}
                    </div>
                )}
            </div>

            <footer className="feature-view__footer">
                <button type="button" className="back-btn" onClick={onBack}>
                    <span className="back-btn__circle">B</span>
                    Back
                </button>
            </footer>
        </div>
    );
}

function formatAnalysis(a: domain.TypeAnalysis): string {
    const list = (items: string[]) =>
        items.length ? items.join(', ') : '—';
    return [
        `**${a.pokemonName}**`,
        `Types: ${list(a.types)}`,
        `Strengths vs: ${list(a.strengths)}`,
        `Weaknesses vs: ${list(a.weaknesses)}`,
        `Resists: ${list(a.resistances)}`,
    ].join('\n\n');
}
