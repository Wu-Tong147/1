import { uiT } from '@/lib/i18n';

function PageLoader() {
    return (
        <div className="grid h-screen w-full place-items-center">
            <p>{uiT('Loading...')}</p>
        </div>
    );
}

export default PageLoader;
