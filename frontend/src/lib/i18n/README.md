# Frontend internationalization

PentAGI uses `i18next` and `react-i18next` for frontend UI localization.

## Adding a language

1. Add a translation resource beside `en-US.ts`.
2. Register it in `index.ts` and extend `SUPPORTED_LANGUAGES`.
3. Keep translation keys structurally identical to `en-US.ts`.
4. Use `useTranslation()` inside React components and call `t('namespace.key')`.

English remains the fallback language. The selected language is stored under
`pentagi.language` in browser local storage, and the root document `lang`
attribute is updated whenever the language changes.

Do not translate user content, terminal output, tool output, provider model
IDs, API field names, or security product names.
