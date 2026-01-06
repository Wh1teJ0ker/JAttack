import { Outlet } from 'react-router-dom';
import { Sidebar } from './Sidebar';

export function MainLayout() {
    return (
        <div className="h-screen bg-background text-foreground overflow-hidden">
            <Sidebar />
            <main className="ml-64 h-full overflow-auto">
                <Outlet />
            </main>
        </div>
    );
}
