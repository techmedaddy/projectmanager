import { Outlet, Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import { LogOut, CheckCircle2 } from 'lucide-react';
import { Button } from '../ui/button';

export function AppLayout() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="min-h-screen flex flex-col text-stone-900 font-sans">
      <header className="sticky top-0 z-20 border-b border-stone-200/70 bg-white/70 backdrop-blur-xl shadow-[0_1px_0_rgba(28,25,23,0.04)]">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <Link to="/" className="flex items-center gap-2.5 text-stone-900 hover:opacity-85 transition-opacity">
            <span className="inline-flex h-8 w-8 items-center justify-center rounded-xl border border-orange-200/80 bg-orange-50 shadow-sm">
              <CheckCircle2 className="w-4.5 h-4.5 text-orange-600" />
            </span>
            <span className="font-semibold text-lg tracking-tight">projectmanager</span>
          </Link>

          <div className="flex items-center gap-2 sm:gap-4">
            <span className="text-sm font-medium text-stone-600 hidden sm:inline-block">
              {user?.name}
            </span>
            <Button variant="ghost" size="sm" onClick={handleLogout} className="text-stone-500 hover:text-stone-900 hover:bg-stone-100">
              <LogOut className="w-4 h-4 mr-2" />
              Logout
            </Button>
          </div>
        </div>
      </header>

      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="rounded-3xl border border-stone-200/70 bg-white/75 p-4 sm:p-6 lg:p-8 shadow-[0_10px_35px_rgba(28,25,23,0.06)] backdrop-blur-sm">
          <Outlet />
        </div>
      </main>
    </div>
  );
}
