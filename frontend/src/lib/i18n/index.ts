import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';

import enUS from './en-US';
import zhCN from './zh-CN';

export const DEFAULT_LANGUAGE = 'en-US';
export const LANGUAGE_STORAGE_KEY = 'pentagi.language';
export const SUPPORTED_LANGUAGES = ['en-US', 'zh-CN'] as const;

export type SupportedLanguage = (typeof SUPPORTED_LANGUAGES)[number];

export const normalizeLanguage = (language?: null | string): SupportedLanguage =>
    language?.toLowerCase().startsWith('zh') ? 'zh-CN' : DEFAULT_LANGUAGE;

export const getInitialLanguage = (): SupportedLanguage => {
    try {
        const storedLanguage = window.localStorage.getItem(LANGUAGE_STORAGE_KEY);

        if (storedLanguage) {
            return normalizeLanguage(storedLanguage);
        }
    } catch {
        // Storage can be disabled by the browser. Fall back to its UI language.
    }

    return normalizeLanguage(typeof navigator === 'undefined' ? undefined : navigator.language);
};

const applyDocumentLanguage = (language: string) => {
    if (typeof document !== 'undefined') {
        document.documentElement.lang = normalizeLanguage(language);
    }
};

void i18n.use(initReactI18next).init({
    fallbackLng: DEFAULT_LANGUAGE,
    initImmediate: false,
    interpolation: {
        escapeValue: false,
    },
    lng: getInitialLanguage(),
    resources: {
        'en-US': { translation: enUS },
        'zh-CN': { translation: zhCN },
    },
    supportedLngs: [...SUPPORTED_LANGUAGES],
});

applyDocumentLanguage(i18n.resolvedLanguage || i18n.language);

i18n.on('languageChanged', (language) => {
    const normalizedLanguage = normalizeLanguage(language);

    applyDocumentLanguage(normalizedLanguage);

    try {
        window.localStorage.setItem(LANGUAGE_STORAGE_KEY, normalizedLanguage);
    } catch {
        // The language still changes for the current session when storage is unavailable.
    }
});

export default i18n;
