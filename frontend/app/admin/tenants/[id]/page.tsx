"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { FiAlertCircle, FiCheck, FiTrash2, FiInfo, FiBarChart2 } from "react-icons/fi";
import Link from "next/link";
import { fetchApi } from "@/lib/api";

import { Button } from "@/components/ui/Button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/Card";
import { Input } from "@/components/ui/Input";
import { Badge } from "@/components/ui/Badge";

interface Tenant {
  id: string;
  name: string;
  slug: string;
  user_count: number;
  run_count: number;
  created_at: string;
}

export default function EditTenantPage() {
  const router = useRouter();
  const params = useParams();
  const tenantId = params.id as string;

  const [tenant, setTenant] = useState<Tenant | null>(null);
  const [name, setName] = useState("");
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    const fetchTenant = async () => {
      try {
        const data = await fetchApi(`/admin/tenants/${tenantId}`);
        setTenant(data);
        setName(data.name);
      } catch (err) {
        console.error("Failed to fetch tenant:", err);
        setError("Failed to load tenant");
      } finally {
        setLoading(false);
      }
    };

    fetchTenant();
  }, [tenantId]);

  const handleUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!name.trim()) {
      setError("Tenant name is required");
      return;
    }

    try {
      setUpdating(true);
      await fetchApi(`/admin/tenants/${tenantId}`, {
        method: "PUT",
        body: JSON.stringify({ name }),
      });

      setSuccess(true);
      setTimeout(() => {
        setSuccess(false);
      }, 2000);

      if (tenant) {
        setTenant({ ...tenant, name });
      }
    } catch (err: any) {
      setError(err.message || "Failed to update tenant");
    } finally {
      setUpdating(false);
    }
  };

  const handleDelete = async () => {
    if (
      !confirm(
        "Are you sure? This will delete the tenant and all associated data."
      )
    ) {
      return;
    }

    try {
      setDeleting(true);
      await fetchApi(`/admin/tenants/${tenantId}`, {
        method: "DELETE",
      });

      router.push("/admin/tenants");
    } catch (err: any) {
      setError(err.message || "Failed to delete tenant");
      setDeleting(false);
    }
  };

  if (loading) return <div className="text-gray-500 text-center py-12">Loading tenant details...</div>;
  if (!tenant) return <div className="text-red-500 text-center py-12">Tenant not found</div>;

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 tracking-tight">Tenant Settings</h1>
          <div className="flex items-center gap-2 mt-1">
            <span className="text-gray-500 text-sm">{tenant.name}</span>
            <Badge variant="default">{tenant.slug}</Badge>
          </div>
        </div>
        <Link href="/admin/tenants">
          <Button variant="secondary" size="sm">Back to List</Button>
        </Link>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="md:col-span-2 space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <FiInfo className="w-4 h-4 text-blue-500" />
                General Information
              </CardTitle>
            </CardHeader>
            <CardContent>
              {success && (
                <div className="mb-6 p-3 bg-green-50 border border-green-200 rounded-lg flex items-center gap-2 text-sm text-green-700">
                  <FiCheck /> Tenant updated successfully!
                </div>
              )}
              {error && (
                <div className="mb-6 p-3 bg-red-50 border border-red-200 rounded-lg flex items-center gap-2 text-sm text-red-700">
                  <FiAlertCircle /> {error}
                </div>
              )}

              <form onSubmit={handleUpdate} className="space-y-4">
                <div className="space-y-2">
                  <label className="text-sm font-semibold text-gray-700">Tenant Name</label>
                  <Input
                    type="text"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="Enter new tenant name"
                  />
                </div>
                <Button 
                  type="submit" 
                  variant="primary" 
                  disabled={updating || name === tenant.name}
                >
                  {updating ? "Saving..." : "Save Changes"}
                </Button>
              </form>
            </CardContent>
          </Card>

          <Card className="border-red-100">
            <CardHeader>
              <CardTitle className="text-lg text-red-600">Danger Zone</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-gray-600 mb-4">
                Deleting this tenant will permanently remove all associated users, workflows, and execution history.
              </p>
              <Button
                variant="danger"
                onClick={handleDelete}
                disabled={deleting}
                className="gap-2"
              >
                <FiTrash2 />
                Delete this Tenant
              </Button>
            </CardContent>
          </Card>
        </div>

        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <FiBarChart2 className="w-4 h-4 text-purple-500" />
                Usage Stats
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="p-4 bg-gray-50 rounded-lg border border-gray-100">
                <p className="text-xs text-gray-500 uppercase font-bold tracking-wider mb-1">Total Users</p>
                <p className="text-2xl font-bold text-gray-900">{tenant.user_count}</p>
              </div>
              <div className="p-4 bg-gray-50 rounded-lg border border-gray-100">
                <p className="text-xs text-gray-500 uppercase font-bold tracking-wider mb-1">Total Workflow Runs</p>
                <p className="text-2xl font-bold text-gray-900">{tenant.run_count}</p>
              </div>
              <div className="pt-2">
                <p className="text-xs text-gray-400">
                  Tenant since {new Date(tenant.created_at).toLocaleDateString()}
                </p>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
