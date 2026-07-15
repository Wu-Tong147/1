import { afterEach, describe, expect, it } from 'vitest';

import { getInitialLanguage, LANGUAGE_STORAGE_KEY, normalizeLanguage } from './index';

afterEach(() => {
    window.localStorage.clear();
});

describe('i18n language selection', () => {
    it('normalizes Chinese browser variants to Simplified Chinese', () => {
        expect(normalizeLanguage('zh')).toBe('zh-CN');
        expect(normalizeLanguage('zh-Hans-CN')).toBe('zh-CN');
    });

    it('falls back to English for unsupported languages', () => {
        expect(normalizeLanguage('de-DE')).toBe('en-US');
    });

    it('prefers a stored language over browser detection', () => {
        window.localStorage.setItem(LANGUAGE_STORAGE_KEY, 'zh-CN');

        expect(getInitialLanguage()).toBe('zh-CN');
    });
});
