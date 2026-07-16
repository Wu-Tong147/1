import { Loader2 } from 'lucide-react';
import { useLocation, useSearchParams } from 'react-router-dom';

import Logo from '@/components/icons/logo';
import { LanguageToggle } from '@/components/shared/language-toggle';
import LoginForm from '@/features/authentication/login-form';
import { getSafeReturnUrl } from '@/lib/utils/auth';
import { useUser } from '@/providers/user-provider';

function Login() {
    const [searchParams] = useSearchParams();
    const location = useLocation();
    const { authInfo, isLoading } = useUser();
    const authProviders = authInfo?.providers || [];

    const returnUrl = getSafeReturnUrl((location.state?.from as string) || searchParams.get('returnUrl'), '/flows/new');

    return (
        <div className="relative flex h-dvh w-full items-center justify-center">
            <LanguageToggle className="absolute top-4 right-4 z-10" />
            <div className="h-dvh w-full lg:grid lg:grid-cols-2">
                <div className="flex items-center justify-center px-4 py-12">
                    {!isLoading ? (
                        <LoginForm
                            providers={authProviders}
                            returnUrl={returnUrl}
                        />
                    ) : (
                        <Loader2 className="size-16 animate-spin" />
                    )}
                </div>
                <div className="from-primary/20 via-primary/10 to-background hidden bg-linear-to-br lg:flex">
                    <Logo className="animate-logo-spin text-foreground m-auto size-32 delay-10000" />
                </div>
            </div>
        </div>
    );
}

export default Login;
