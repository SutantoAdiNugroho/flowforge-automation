"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import {
  FiPlus,
  FiMail,
  FiLock,
  FiAlertCircle,
  FiCheck,
} from "react-icons/fi";
import Link from "next/link";
import { fetchApi } from "@/lib/api";

import { Button } from "@/components/ui/Button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/Card";
import { Input } from "@/components/ui/Input";

export default function CreateTenantPage() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [adminEmail, setAdminEmail] = useState("");
  const [adminPassword, setAdminPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);

  const generateSlug = (text: string) => {
    return text
      .toLowerCase()
      .replace(/[^a-z0-9-]/g, "-")
      .replace(/--+/g, "-")
      .replace(/^-+|-+$/g, "");
  };

  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newName = e.target.value;
    setName(newName);
    if (!slug) {
      setSlug(generateSlug(newName));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!name || !slug || !adminEmail || !adminPassword) {
      setError("All fields are required");
      return;
    }

    if (adminPassword.length < 6) {
      setError("Password must be at least 6 characters");
      return;
    }

    try {
      setLoading(true);
      await fetchApi("/admin/tenants", {
        method: "POST",
        body: JSON.stringify({
          name,
          slug,
          admin_email: adminEmail,
          admin_password: adminPassword,
        }),
      });

      setSuccess(true);
      setTimeout(() => {
        router.refresh();
        router.push("/admin/tenants");
      }, 1000);
    } catch (err: any) {
      setError(err.message || "Failed to create tenant");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900 tracking-tight">Create Tenant</h1>
        <p className="text-gray-500 text-sm mt-1">Set up a new organization with an initial administrator</p>
      </div>

      <Card>
        <CardContent className="p-6">
          {success && (
            <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded-lg flex items-center gap-3">
              <FiCheck className="w-5 h-5 text-green-600" />
              <span className="text-sm text-green-900">Tenant created successfully! Redirecting...</span>
            </div>
          )}

          {error && (
            <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg flex items-center gap-3">
              <FiAlertCircle className="w-5 h-5 text-red-600" />
              <span className="text-sm text-red-900">{error}</span>
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="space-y-2">
              <label className="text-sm font-semibold text-gray-700 flex items-center gap-2">
                <FiPlus className="w-4 h-4" />
                Tenant Name
              </label>
              <Input
                type="text"
                value={name}
                onChange={handleNameChange}
                placeholder="e.g. Acme Corp"
                required
              />
            </div>

            <div className="space-y-2">
              <label className="text-sm font-semibold text-gray-700">Slug</label>
              <Input
                type="text"
                value={slug}
                onChange={(e) => setSlug(e.target.value)}
                placeholder="e.g. acme-corp"
                required
              />
              <p className="text-xs text-gray-400">Used for subdomains and URLs (lowercase, hyphens only)</p>
            </div>

            <hr className="border-gray-100 my-6" />
            <h3 className="text-sm font-bold text-gray-900 uppercase tracking-wider">Initial Admin Credentials</h3>

            <div className="space-y-2">
              <label className="text-sm font-semibold text-gray-700 flex items-center gap-2">
                <FiMail className="w-4 h-4" />
                Email Address
              </label>
              <Input
                type="email"
                value={adminEmail}
                onChange={(e) => setAdminEmail(e.target.value)}
                placeholder="admin@acme.com"
                required
              />
            </div>

            <div className="space-y-2">
              <label className="text-sm font-semibold text-gray-700 flex items-center gap-2">
                <FiLock className="w-4 h-4" />
                Password
              </label>
              <Input
                type="password"
                value={adminPassword}
                onChange={(e) => setAdminPassword(e.target.value)}
                placeholder="••••••••"
                required
              />
            </div>

            <div className="flex gap-3 pt-4">
              <Button
                type="submit"
                variant="primary"
                className="flex-1"
                disabled={loading}
              >
                {loading ? "Creating..." : "Create Tenant"}
              </Button>
              <Link href="/admin/tenants" className="flex-1">
                <Button variant="secondary" className="w-full">
                  Cancel
                </Button>
              </Link>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
