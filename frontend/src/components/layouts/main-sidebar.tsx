import { Avatar, AvatarFallback } from '@radix-ui/react-avatar';
import {
    ChevronsUpDown,
    Clock,
    FileText,
    Folder,
    GitFork,
    KeyRound,
    Languages,
    LayoutDashboard,
    LibraryBig,
    LogOut,
    Monitor,
    Moon,
    Plus,
    Settings,
    Settings2,
    Star,
    Sun,
    UserIcon,
} from 'lucide-react';
import { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useMatch, useParams } from 'react-router-dom';

import type { Flow } from '@/providers/sidebar-flows-provider';
import type { Theme } from '@/providers/theme-provider';

import Logo from '@/components/icons/logo';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
    Sidebar,
    SidebarContent,
    SidebarFooter,
    SidebarGroup,
    SidebarGroupContent,
    SidebarGroupLabel,
    SidebarHeader,
    SidebarMenu,
    SidebarMenuAction,
    SidebarMenuButton,
    SidebarMenuItem,
    SidebarRail,
} from '@/components/ui/sidebar';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { PasswordChangeForm } from '@/features/authentication/password-change-form';
import { useResourcesUpload } from '@/features/resources/use-resources-upload';
import { useTheme } from '@/hooks/use-theme';
import { normalizeLanguage } from '@/lib/i18n';
import { useFavorites } from '@/providers/favorites-provider';
import { useSidebarFlows } from '@/providers/sidebar-flows-provider';
import { useUser } from '@/providers/user-provider';

interface FlowMenuItemProps {
    activeFlowId: null | number;
    flow: Flow;
    isFavorite: boolean;
    onToggleFavorite: (flowId: string) => void;
}

export function MainSidebar() {
    const { i18n, t } = useTranslation();
    const [isPasswordModalOpen, setIsPasswordModalOpen] = useState(false);
    const isDashboardActive = useMatch('/dashboard');
    const isFlowsActive = useMatch('/flows/*');
    const isTemplatesActive = useMatch('/templates/*');
    const isKnowledgesActive = useMatch('/knowledges/*');
    const isResourcesActive = useMatch('/resources/*');
    const isSettingsActive = useMatch('/settings/*');
    const { flowId: flowIdParam } = useParams<{ flowId: string }>();

    const { authInfo, logout } = useUser();
    const user = authInfo?.user;
    const { setTheme, theme } = useTheme();
    const { addFavoriteFlow, favoriteFlowIds, removeFavoriteFlow } = useFavorites();
    const { flows } = useSidebarFlows();

    const resourcesUpload = useResourcesUpload();

    const currentLanguage = normalizeLanguage(i18n.resolvedLanguage || i18n.language);
    const nextLanguage = currentLanguage === 'zh-CN' ? 'en-US' : 'zh-CN';

    const flowId = useMemo(() => (flowIdParam ? Number(flowIdParam) : null), [flowIdParam]);

    const favoriteFlows = useMemo(
        () =>
            flows
                .filter((flow) => favoriteFlowIds.includes(Number(flow.id)))
                .sort((a, b) => Number(b.id) - Number(a.id)),
        [flows, favoriteFlowIds],
    );

    const recentFlows = useMemo(
        () =>
            flows
                .filter((flow) => !favoriteFlowIds.includes(Number(flow.id)))
                .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
                .slice(0, 5),
        [flows, favoriteFlowIds],
    );

    return (
        <Sidebar collapsible="icon">
            <SidebarHeader>
                <SidebarMenu>
                    <SidebarMenuItem className="flex items-center gap-2">
                        <div className="flex aspect-square size-8 items-center justify-center">
                            <Logo className="hover:animate-logo-spin size-6" />
                        </div>
                        <div className="grid flex-1 text-left leading-tight">
                            <span className="truncate font-semibold">PentAGI</span>
                        </div>
                    </SidebarMenuItem>
                </SidebarMenu>
            </SidebarHeader>
            <SidebarContent>
                <SidebarGroup className="bg-sidebar sticky top-0 z-10">
                    <SidebarGroupContent>
                        <SidebarMenu>
                            <SidebarMenuItem className="group-data-[state=expanded]:hidden">
                                <SidebarMenuButton asChild>
                                    <Link to="/flows/new">
                                        <Plus />
                                        {t('navigation.newFlow')}
                                    </Link>
                                </SidebarMenuButton>
                            </SidebarMenuItem>
                            <SidebarMenuItem>
                                <SidebarMenuButton
                                    asChild
                                    isActive={!!isDashboardActive}
                                >
                                    <Link to="/dashboard">
                                        <LayoutDashboard />
                                        {t('navigation.dashboard')}
                                    </Link>
                                </SidebarMenuButton>
                            </SidebarMenuItem>
                            <SidebarMenuItem>
                                <SidebarMenuButton
                                    asChild
                                    isActive={!!isFlowsActive}
                                >
                                    <Link to="/flows">
                                        <GitFork />
                                        {t('navigation.flows')}
                                    </Link>
                                </SidebarMenuButton>
                                <SidebarMenuAction
                                    asChild
                                    className="data-[state=open]:bg-accent rounded-sm"
                                    showOnHover
                                >
                                    <Link to="/flows/new">
                                        <Plus />
                                    </Link>
                                </SidebarMenuAction>
                            </SidebarMenuItem>
                            <SidebarMenuItem>
                                <SidebarMenuButton
                                    asChild
                                    isActive={!!isTemplatesActive}
                                >
                                    <Link to="/templates">
                                        <FileText />
                                        {t('navigation.templates')}
                                    </Link>
                                </SidebarMenuButton>
                                <SidebarMenuAction
                                    asChild
                                    className="data-[state=open]:bg-accent rounded-sm"
                                    showOnHover
                                >
                                    <Link to="/templates/new">
                                        <Plus />
                                    </Link>
                                </SidebarMenuAction>
                            </SidebarMenuItem>
                            <SidebarMenuItem>
                                <SidebarMenuButton
                                    asChild
                                    isActive={!!isResourcesActive}
                                >
                                    <Link to="/resources">
                                        <Folder />
                                        {t('navigation.resources')}
                                    </Link>
                                </SidebarMenuButton>
                                <SidebarMenuAction
                                    className="data-[state=open]:bg-accent rounded-sm"
                                    onClick={resourcesUpload.openFilePicker}
                                    showOnHover
                                    title={t('navigation.uploadFile')}
                                    type="button"
                                >
                                    <Plus />
                                </SidebarMenuAction>
                            </SidebarMenuItem>
                            <SidebarMenuItem>
                                <SidebarMenuButton
                                    asChild
                                    isActive={!!isKnowledgesActive}
                                >
                                    <Link to="/knowledges">
                                        <LibraryBig />
                                        {t('navigation.knowledges')}
                                    </Link>
                                </SidebarMenuButton>
                                <SidebarMenuAction
                                    asChild
                                    className="data-[state=open]:bg-accent rounded-sm"
                                    showOnHover
                                >
                                    <Link to="/knowledges/new">
                                        <Plus />
                                    </Link>
                                </SidebarMenuAction>
                            </SidebarMenuItem>
                        </SidebarMenu>
                    </SidebarGroupContent>
                </SidebarGroup>

                {recentFlows.length > 0 && (
                    <SidebarGroup>
                        <SidebarGroupLabel className="flex items-center gap-2">
                            <Clock />
                            {t('navigation.recentFlows')}
                        </SidebarGroupLabel>
                        <SidebarGroupContent>
                            <SidebarMenu>
                                {recentFlows.map((flow) => (
                                    <FlowMenuItem
                                        activeFlowId={flowId}
                                        flow={flow}
                                        isFavorite={false}
                                        key={flow.id}
                                        onToggleFavorite={addFavoriteFlow}
                                    />
                                ))}
                            </SidebarMenu>
                        </SidebarGroupContent>
                    </SidebarGroup>
                )}

                {favoriteFlows.length > 0 && (
                    <SidebarGroup>
                        <SidebarGroupLabel className="flex items-center gap-2">
                            <Star />
                            {t('navigation.favoriteFlows')}
                        </SidebarGroupLabel>
                        <SidebarGroupContent>
                            <SidebarMenu>
                                {favoriteFlows.map((flow) => (
                                    <FlowMenuItem
                                        activeFlowId={flowId}
                                        flow={flow}
                                        isFavorite
                                        key={flow.id}
                                        onToggleFavorite={removeFavoriteFlow}
                                    />
                                ))}
                            </SidebarMenu>
                        </SidebarGroupContent>
                    </SidebarGroup>
                )}
            </SidebarContent>
            <SidebarFooter>
                <SidebarMenu>
                    <SidebarMenuItem>
                        <SidebarMenuButton
                            asChild
                            isActive={!!isSettingsActive}
                        >
                            <Link to="/settings">
                                <Settings />
                                {t('navigation.settings')}
                            </Link>
                        </SidebarMenuButton>
                    </SidebarMenuItem>
                    <SidebarMenuItem>
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <SidebarMenuButton
                                    className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
                                    size="lg"
                                >
                                    <Avatar className="bg-background dark:bg-muted size-8 rounded-lg">
                                        <AvatarFallback className="flex size-8 items-center justify-center">
                                            <UserIcon className="size-4" />
                                        </AvatarFallback>
                                    </Avatar>
                                    <div className="grid flex-1 text-left text-sm leading-tight">
                                        <span className="truncate font-semibold">{user?.name}</span>
                                        <span className="truncate text-xs">{user?.mail}</span>
                                    </div>
                                    <ChevronsUpDown className="ml-auto size-4" />
                                </SidebarMenuButton>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent
                                align="end"
                                className="w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg"
                                side="bottom"
                                sideOffset={4}
                            >
                                <DropdownMenuLabel className="p-0 font-normal">
                                    <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
                                        <Avatar className="bg-muted flex size-8 items-center justify-center rounded-lg">
                                            <AvatarFallback className="flex items-center justify-center rounded-lg">
                                                <UserIcon className="size-4" />
                                            </AvatarFallback>
                                        </Avatar>
                                        <div className="grid flex-1 text-left text-sm leading-tight">
                                            <span className="truncate font-semibold">{user?.name}</span>
                                            <span className="truncate text-xs">{user?.mail}</span>
                                            <span className="text-muted-foreground truncate text-xs">
                                                {user?.type === 'local' ? 'local' : 'oauth'}
                                            </span>
                                        </div>
                                    </div>
                                </DropdownMenuLabel>
                                <DropdownMenuSeparator />
                                <DropdownMenuItem
                                    className="cursor-default hover:bg-transparent focus:bg-transparent"
                                    onSelect={(event) => event.preventDefault()}
                                >
                                    <Settings2 />
                                    {t('navigation.theme')}
                                    <Tabs
                                        className="-my-1.5 -mr-2 ml-auto"
                                        onValueChange={(value) => setTheme(value as Theme)}
                                        value={theme || 'system'}
                                    >
                                        <TabsList className="h-7 p-0.5">
                                            <TabsTrigger
                                                aria-label={t('navigation.themeSystem')}
                                                className="h-6 px-2"
                                                value="system"
                                            >
                                                <Monitor className="size-4" />
                                            </TabsTrigger>
                                            <TabsTrigger
                                                aria-label={t('navigation.themeLight')}
                                                className="h-6 px-2"
                                                value="light"
                                            >
                                                <Sun className="size-4" />
                                            </TabsTrigger>
                                            <TabsTrigger
                                                aria-label={t('navigation.themeDark')}
                                                className="h-6 px-2"
                                                value="dark"
                                            >
                                                <Moon className="size-4" />
                                            </TabsTrigger>
                                        </TabsList>
                                    </Tabs>
                                </DropdownMenuItem>
                                <DropdownMenuItem onClick={() => void i18n.changeLanguage(nextLanguage)}>
                                    <Languages className="mr-2 size-4" />
                                    {t('language.label')}
                                    <span className="text-muted-foreground ml-auto text-xs">
                                        {t('language.action')}
                                    </span>
                                </DropdownMenuItem>
                                {user?.type === 'local' && (
                                    <>
                                        <DropdownMenuSeparator />
                                        <DropdownMenuItem onClick={() => setIsPasswordModalOpen(true)}>
                                            <KeyRound className="mr-2 size-4" />
                                            {t('navigation.changePassword')}
                                        </DropdownMenuItem>
                                    </>
                                )}
                                <DropdownMenuSeparator />
                                <DropdownMenuItem onClick={() => logout()}>
                                    <LogOut className="mr-2 size-4" />
                                    {t('navigation.logOut')}
                                </DropdownMenuItem>
                            </DropdownMenuContent>
                        </DropdownMenu>
                    </SidebarMenuItem>
                </SidebarMenu>
            </SidebarFooter>
            <SidebarRail />

            <input
                aria-hidden="true"
                className="hidden"
                key={resourcesUpload.fileInputKey}
                multiple
                name="resource-upload"
                tabIndex={-1}
                type="file"
                {...resourcesUpload.fileInputProps}
            />

            <Dialog
                onOpenChange={setIsPasswordModalOpen}
                open={isPasswordModalOpen}
            >
                <DialogContent className="sm:max-w-[425px]">
                    <DialogHeader>
                        <DialogTitle>{t('navigation.changePassword')}</DialogTitle>
                    </DialogHeader>
                    <PasswordChangeForm
                        onCancel={() => setIsPasswordModalOpen(false)}
                        onSuccess={() => setIsPasswordModalOpen(false)}
                    />
                </DialogContent>
            </Dialog>
        </Sidebar>
    );
}

function FlowMenuItem({ activeFlowId, flow, isFavorite, onToggleFavorite }: FlowMenuItemProps) {
    const { t } = useTranslation();

    return (
        <SidebarMenuItem>
            <SidebarMenuButton
                asChild
                isActive={activeFlowId === Number(flow.id)}
            >
                <Link to={`/flows/${flow.id}`}>
                    <span className="-mx-2 w-8 shrink-0 text-center text-xs group-data-[state=expanded]:hidden">
                        {flow.id}
                    </span>
                    <span className="text-muted-foreground bg-background dark:bg-muted -my-0.5 -ml-0.5 h-5 min-w-5 shrink-0 rounded-md px-px py-0.5 text-center text-xs group-data-[state=collapsed]:hidden">
                        {flow.id}
                    </span>
                    <span className="truncate">{flow.title}</span>
                </Link>
            </SidebarMenuButton>
            <SidebarMenuAction
                aria-label={t('navigation.toggleFavorite')}
                aria-pressed={isFavorite}
                className="data-[state=open]:bg-accent rounded-sm"
                onClick={() => onToggleFavorite(flow.id)}
                showOnHover
            >
                <Star className={isFavorite ? 'fill-yellow-500 stroke-yellow-500' : ''} />
            </SidebarMenuAction>
        </SidebarMenuItem>
    );
}
