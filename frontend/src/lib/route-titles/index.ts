import type { ComponentType } from 'react';

import {
    useFlowQuery,
    useFlowTemplateQuery,
    useKnowledgeDocumentQuery,
    useSettingsProvidersQuery,
} from '@/graphql/types';
import { uiT } from '@/lib/i18n';

import { apolloTitle } from './apollo-title';
import { formatPromptId } from './format-prompt-id';
import { type RouteParams } from './render-title';

export interface RouteTitleHandle {
    title: TitleResolver;
}

/**
 * A `handle.title` value can be one of three forms:
 *   - `string` — fully static, known at build time.
 *   - `(params) => string` — derived synchronously from URL params.
 *   - `ComponentType<{ params }>` — reactive (e.g. subscribes to Apollo
 *     cache for resource-driven titles). Must be produced by `apolloTitle()`
 *     so the marker it attaches lets `DocumentTitle` distinguish a component
 *     from a `(params) => string` resolver at runtime. A hand-rolled component
 *     function will be misdetected as a resolver and called with raw params —
 *     always route reactive titles through `apolloTitle()`.
 */
export type TitleResolver = ((params: RouteParams) => string) | ComponentType<{ params: RouteParams }> | string;

/**
 * Single source of truth for every route's document `<title>`. `app.tsx`
 * imports nothing from Apollo for title purposes — it only spreads handles
 * from this registry onto the matching <Route>.
 */
export const routeTitles = {
    apiTokens: { title: uiT('API Tokens') },
    dashboard: { title: uiT('Dashboard') },
    flow: {
        title: apolloTitle({
            select: (data, { flowId }) =>
                data?.flow?.title && flowId ? `Flow #${flowId} — ${data.flow.title}` : 'Flow',
            useQuery: useFlowQuery,
            variables: ({ flowId }) => (flowId ? { id: flowId } : null),
        }),
    },
    flowReport: { title: uiT('Flow report') },
    flows: { title: uiT('Flows') },
    knowledge: {
        title: apolloTitle({
            select: (data, { knowledgeId }) =>
                knowledgeId === 'new' ? uiT('New knowledge') : data?.knowledgeDocument?.question || uiT('Knowledge'),
            useQuery: useKnowledgeDocumentQuery,
            variables: ({ knowledgeId }) => (!knowledgeId || knowledgeId === 'new' ? null : { id: knowledgeId }),
        }),
    },
    knowledges: { title: uiT('Knowledges') },
    login: { title: 'Login' },
    newFlow: { title: uiT('New flow') },
    oauth: { title: 'OAuth' },
    prompt: {
        title: (params: RouteParams) => (params.promptId ? formatPromptId(params.promptId) : 'Prompt'),
    },
    prompts: { title: uiT('Prompts') },

    provider: {
        title: apolloTitle({
            select: (data, { providerId }) => {
                if (providerId === 'new') {
                    return uiT('New provider');
                }

                const provider = data?.settingsProviders.userDefined?.find(
                    (candidate) => String(candidate.id) === providerId,
                );

                return provider?.name || uiT('Provider');
            },
            useQuery: useSettingsProvidersQuery,
            variables: ({ providerId }) => (providerId === 'new' ? null : {}),
        }),
    },

    providers: { title: uiT('Providers') },

    resources: { title: uiT('Resources') },

    template: {
        title: apolloTitle({
            select: (data, { templateId }) =>
                templateId === 'new' ? uiT('New template') : data?.flowTemplate?.title || uiT('Template'),
            useQuery: useFlowTemplateQuery,
            variables: ({ templateId }) => (!templateId || templateId === 'new' ? null : { templateId }),
        }),
    },

    templates: { title: uiT('Templates') },
} as const satisfies Record<string, RouteTitleHandle>;
