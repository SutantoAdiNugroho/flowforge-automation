"use client";

import { useAuth } from "@/contexts/AuthContext";
import { useRouter, usePathname } from "next/navigation";
import { useEffect, useState } from "react";
import { FiLogOut, FiMenu, FiX, FiShield, FiBriefcase } from "react-icons/fi";
import Link from "next/link";
import { Button } from "@/components/ui/Button";

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { user, logout } = useAuth();
  const router = useRouter();
  const pathname = usePathname();
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    if (!mounted) return;

    if (!user || user.role !== "super-admin") {
      router.push("/dashboard");
    }
  }, [user, mounted, router]);

  const isActive = (path: string) => pathname === path || pathname.startsWith(path + "/");

  if (!mounted || !user || user.role !== "super-admin") {
    return null;
  }

  return (
    <div className="flex h-screen bg-gray-100">
      <aside
        className={`${
          sidebarOpen ? "w-64" : "w-20"
        } bg-gray-900 text-white transition-all duration-300 flex flex-col`}
      >
        <div className="p-4 flex items-center justify-between border-b border-gray-800">
          {sidebarOpen && (
            <div className="flex items-center gap-2">
              <FiShield className="w-6 h-6 text-red-400" />
              <h1 className="text-xl font-bold">Admin Panel</h1>
            </div>
          )}
          <button
            onClick={() => setSidebarOpen(!sidebarOpen)}
            className="p-2 hover:bg-gray-800 rounded-lg transition-colors"
          >
            {sidebarOpen ? <FiX /> : <FiMenu />}
          </button>
        </div>

        <nav className="flex-1 px-4 py-6 space-y-2">
          <Link
            href="/admin/tenants"
            className={`flex items-center gap-3 p-3 rounded-lg transition-colors hover:bg-gray-800 ${
              isActive("/admin/tenants") ? "bg-gray-800" : ""
            }`}
            title="Tenants"
          >
            <FiBriefcase className="w-5 h-5 flex-shrink-0" />
            {sidebarOpen && <span>Tenants</span>}
          </Link>
          
        </nav>

        <div className="p-4 border-t border-gray-800">
          <Button
            variant="danger"
            onClick={logout}
            className="w-full gap-2"
          >
            <FiLogOut />
            {sidebarOpen && <span>Logout</span>}
          </Button>
        </div>
      </aside>

      <main className="flex-1 overflow-auto p-8">
        {children}
      </main>
    </div>
  );
}
