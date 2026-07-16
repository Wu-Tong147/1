import { afterEach, describe, expect, it } from 'vitest';

import i18n, { getInitialLanguage, LANGUAGE_STORAGE_KEY, normalizeLanguage, uiT } from './index';
import { enUSUI, zhCNUI } from './ui';

afterEach(async () => {
    window.localStorage.clear();
    await i18n.changeLanguage('en-US');
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

describe('UI translation catalog', () => {
    it('keeps the English and Simplified Chinese catalogs structurally identical', () => {
        const englishKeys = Object.keys(enUSUI).sort();
        const chineseKeys = Object.keys(zhCNUI).sort();

        expect(chineseKeys.length).toBeGreaterThan(400);
        expect(englishKeys).toEqual(chineseKeys);
        expect(Object.values(zhCNUI).every((value) => value.trim().length > 0)).toBe(true);
    });

    it('resolves shared UI text in the active language', async () => {
        await i18n.changeLanguage('zh-CN');
        expect(uiT('Settings')).toBe('设置');
        expect(document.documentElement.lang).toBe('zh-CN');

        await i18n.changeLanguage('en-US');
        expect(uiT('Settings')).toBe('Settings');
        expect(document.documentElement.lang).toBe('en-US');
    });

    it('falls back to the source text for a missing UI entry', () => {
        expect(uiT('Unregistered upstream text')).toBe('Unregistered upstream text');
    });
});
