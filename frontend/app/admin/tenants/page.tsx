"use client";

import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { FiPlus, FiEdit2, FiTrash2, FiBriefcase } from "react-icons/fi";
import Link from "next/link";
import { fetchApi } from "@/lib/api";

import { Button } from "@/components/ui/Button";
import { Badge } from "@/components/ui/Badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/Card";
import { Table, TableBody, TableCell, TableHead, TableHeader as UITableHeader, TableRow } from "@/components/ui/Table";

interface Tenant {
  id: string;
  name: string;
  slug: string;
  user_count: number;
  run_count: number;
  created_at: string;
}

interface PaginationResponse {
  content: Tenant[];
  pagination: {
    total: number;
    page: number;
    page_size: number;
  };
}

export default function AdminTenantsPage() {
  const router = useRouter();
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [deleting, setDeleting] = useState<string | null>(null);
  const hasInitialized = useRef(false);

  const pageSize = 50;

  const fetchTenants = async () => {
    try {
      setLoading(true);
      const data: PaginationResponse = await fetchApi(
        `/admin/tenants?page=${page}&page_size=${pageSize}`
      );
      setTenants(data.content || []);
      if (data.pagination) {
        setTotal(data.pagination.total);
      }
    } catch (err) {
      console.error("Failed to fetch tenants:", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (hasInitialized.current) return;
    hasInitialized.current = true;
    fetchTenants();
  }, []);

  useEffect(() => {
    if (!hasInitialized.current) return;
    fetchTenants();
  }, [page]);

  const handleDelete = async (id: string) => {
    if (!confirm("Are you sure you want to delete this tenant?")) return;

    try {
      setDeleting(id);
      await fetchApi(`/admin/tenants/${id}`, { method: "DELETE" });
      setTenants(tenants.filter((t) => t.id !== id));
      setTotal(total - 1);
    } catch (err) {
      console.error("Failed to delete tenant:", err);
    } finally {
      setDeleting(null);
    }
  };

  const totalPages = Math.ceil(total / pageSize);

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 tracking-tight">Tenants</h1>
          <p className="text-gray-500 text-sm">Manage organizations and their resources</p>
        </div>
        <Link href="/admin/tenants/new">
          <Button variant="primary" className="gap-2">
            <FiPlus />
            New Tenant
          </Button>
        </Link>
      </div>

      <Card>
        <CardContent className="p-0">
          <Table>
            <UITableHeader>
              <TableRow>
                <TableHead>Tenant Name</TableHead>
                <TableHead>Slug</TableHead>
                <TableHead>Users</TableHead>
                <TableHead>Total Runs</TableHead>
                <TableHead>Created At</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </UITableHeader>
            <TableBody>
              {loading && tenants.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-center py-12 text-gray-500">
                    Loading tenants...
                  </TableCell>
                </TableRow>
              ) : tenants.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-center py-12 text-gray-500">
                    No tenants found
                  </TableCell>
                </TableRow>
              ) : (
                tenants.map((tenant) => (
                  <TableRow key={tenant.id}>
                    <TableCell className="font-medium text-gray-900">
                      {tenant.name}
                    </TableCell>
                    <TableCell>
                      <Badge variant="default">{tenant.slug}</Badge>
                    </TableCell>
                    <TableCell className="text-gray-600">
                      {tenant.user_count}
                    </TableCell>
                    <TableCell className="text-gray-600">
                      {tenant.run_count}
                    </TableCell>
                    <TableCell className="text-gray-500 text-sm">
                      {new Date(tenant.created_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Link href={`/admin/tenants/${tenant.id}`}>
                          <Button variant="ghost" size="sm" className="text-blue-600">
                            <FiEdit2 className="w-4 h-4" />
                          </Button>
                        </Link>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="text-red-600"
                          onClick={() => handleDelete(tenant.id)}
                          disabled={deleting === tenant.id}
                        >
                          <FiTrash2 className="w-4 h-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {totalPages > 1 && (
        <div className="flex justify-center items-center gap-4">
          <Button
            variant="secondary"
            size="sm"
            onClick={() => setPage(Math.max(1, page - 1))}
            disabled={page === 1}
          >
            Previous
          </Button>
          <span className="text-sm text-gray-600 font-medium">
            Page {page} of {totalPages}
          </span>
          <Button
            variant="secondary"
            size="sm"
            onClick={() => setPage(Math.min(totalPages, page + 1))}
            disabled={page === totalPages}
          >
            Next
          </Button>
        </div>
      )}
    </div>
  );
}
