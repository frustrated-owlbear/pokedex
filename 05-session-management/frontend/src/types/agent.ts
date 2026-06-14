export type StepKind =
    | 'Event'
    | 'Thought'
    | 'Action'
    | 'Observation'
    | 'FinalAnswer';

export type TraceStep = {
    id: string;
    kind: StepKind;
    timestamp: string;
    title: string;
    detail?: string;
    tools?: string[];
};

export type TraceFilter = 'all' | 'events' | 'actions' | 'thoughts' | 'tools';

export type SituationView = {
    location: string;
    region: string;
    time: string;
    period: string;
    weather: string;
    memory: string;
    tools: string[];
    analysis?: string;
};

export type SessionView = {
    id: string;
    startedAt: string;
    endedAt?: string;
    summary?: string;
    eventCount: number;
    active: boolean;
};

export function parseTraceStep(payload: unknown): TraceStep | null {
    if (!payload || typeof payload !== 'object') {
        return null;
    }
    const record = payload as Record<string, unknown>;
    const kind = record.kind;
    if (typeof kind !== 'string') {
        return null;
    }
    return {
        id: typeof record.id === 'string' ? record.id : crypto.randomUUID(),
        kind: kind as StepKind,
        timestamp:
            typeof record.timestamp === 'string'
                ? record.timestamp
                : new Date().toISOString(),
        title: typeof record.title === 'string' ? record.title : 'Step',
        detail: typeof record.detail === 'string' ? record.detail : undefined,
        tools: Array.isArray(record.tools)
            ? record.tools.filter((t): t is string => typeof t === 'string')
            : undefined,
    };
}
