import { Languages } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { Button } from '@/components/ui/button';
import { normalizeLanguage } from '@/lib/i18n';
import { cn } from '@/lib/utils';

interface LanguageToggleProps {
    className?: string;
}

export function LanguageToggle({ className }: LanguageToggleProps) {
    const { i18n, t } = useTranslation();
    const currentLanguage = normalizeLanguage(i18n.resolvedLanguage || i18n.language);
    const nextLanguage = currentLanguage === 'zh-CN' ? 'en-US' : 'zh-CN';

    return (
        <Button
            aria-label={t('language.switchAriaLabel')}
            className={cn('gap-2', className)}
            onClick={() => void i18n.changeLanguage(nextLanguage)}
            size="sm"
            type="button"
            variant="ghost"
        >
            <Languages className="size-4" />
            {t('language.action')}
        </Button>
    );
}
