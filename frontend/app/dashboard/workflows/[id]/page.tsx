"use client";

import { useEffect, useState, use } from "react";
import { fetchApi } from "@/lib/api";
import { useAuth } from "@/contexts/AuthContext";
import { useRouter } from "next/navigation";
import Link from "next/link";

import { Button } from "@/components/ui/Button";
import { Badge } from "@/components/ui/Badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/Card";
import { Table, TableBody, TableCell, TableHead, TableHeader as UITableHeader, TableRow } from "@/components/ui/Table";
import { useSSE } from "@/hooks/useSSE";

export default function WorkflowDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const resolvedParams = use(params);
  const id = resolvedParams.id;
  const router = useRouter();
  const [workflow, setWorkflow] = useState<any>(null);
  const [runs, setRuns] = useState<any[]>([]);
  const [versions, setVersions] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [activatingVersion, setActivatingVersion] = useState<number | null>(null);
  const { hasRole } = useAuth();

  const fetchData = async () => {
    try {
      const [wfData, runsData, versionsData] = await Promise.all([
        fetchApi(`/workflows/${id}`),
        fetchApi(`/workflows/${id}/runs?page=1&page_size=10`),
        fetchApi(`/workflows/${id}/versions?page=1&page_size=100`)
      ]);
      setWorkflow(wfData);
      setRuns(runsData.content || []);
      setVersions(versionsData.content || []);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [id]);

  useSSE("/events", (event) => {
    if (event.type === "run_update") {
      setRuns((prev) => {
        const idx = prev.findIndex((r) => r.id === event.payload.run_id);
        if (idx !== -1) {
          const newRuns = [...prev];
          newRuns[idx] = { ...newRuns[idx], status: event.payload.status };
          return newRuns;
        } else {
          // A new run might have been created, let's just refetch
          fetchData();
          return prev;
        }
      });
    }
  });

  const handleTrigger = async () => {
    try {
      const result = await fetchApi(`/workflows/${id}/runs`, { method: "POST", body: "{}" });
      alert("Workflow berhasil dijalankan!");
      router.push(`/dashboard/runs/${result.id}`);
    } catch (err: any) {
      alert("Gagal menjalankan workflow: " + err.message);
    }
  };

  const handleToggleStatus = async () => {
    const newStatus = !workflow.is_active;
    try {
      await fetchApi(`/workflows/${id}`, {
        method: "PUT",
        body: JSON.stringify({
          ...workflow,
          is_active: newStatus,
          name: workflow.name,
          trigger_type: workflow.trigger_type,
          definition: workflow.definition,
          cron_expression: workflow.cron_expression
        }),
      });
      setWorkflow({ ...workflow, is_active: newStatus });
      router.refresh();
    } catch (err: any) {
      alert("Gagal mengubah status: " + err.message);
    }
  };

  const handleActivateVersion = async (version: number) => {
    try {
      setActivatingVersion(version);
      const updated = await fetchApi(`/workflows/${id}/versions/${version}/activate`, {
        method: "PUT",
      });
      setWorkflow(updated);
      await fetchData();
    } catch (err: any) {
      alert("Gagal mengaktifkan version: " + err.message);
    } finally {
      setActivatingVersion(null);
    }
  };

  if (loading) return <div className="text-gray-500">Loading...</div>;
  if (!workflow) return <div className="text-red-500">Workflow not found</div>;

  const canEdit = hasRole(["admin", "editor"]);

  return (
    <div className="space-y-6">
      <Card>
        <CardContent className="p-6">
          <div className="flex justify-between items-start">
            <div>
              <h1 className="text-2xl font-bold text-gray-900 tracking-tight">{workflow.name}</h1>
              <p className="text-gray-500 mt-1">{workflow.description}</p>
              <div className="mt-4 flex space-x-3 text-sm text-gray-600">
                <Badge variant="default">Type: {workflow.trigger_type}</Badge>
                <Badge variant="default">Version: {workflow.version}</Badge>
                <Badge variant={workflow.is_active ? "success" : "danger"}>
                  {workflow.is_active ? "Active" : "Inactive"}
                </Badge>
              </div>
            </div>
            <div className="flex space-x-2">
              {canEdit && (
                <>
                  <Link href={`/dashboard/workflows/${id}/versions/new`}>
                    <Button variant="primary">New Version</Button>
                  </Link>
                  <Button
                    variant={workflow.is_active ? "danger" : "success"}
                    onClick={handleToggleStatus}
                  >
                    {workflow.is_active ? "Deactivate" : "Activate"}
                  </Button>
                </>
              )}
              <Link href="/dashboard">
                <Button variant="secondary">Back</Button>
              </Link>
            </div>
          </div>

          <div className="mt-8 pt-6 border-t border-gray-100">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-lg font-semibold text-gray-900">Definition (JSON)</h2>
              {canEdit && (
                <Button 
                  variant="primary" 
                  size="sm"
                  className="bg-green-600 hover:bg-green-700 disabled:opacity-50" 
                  onClick={handleTrigger}
                  disabled={!workflow.is_active}
                >
                  Run Workflow
                </Button>
              )}
            </div>

            {workflow.trigger_type === 'webhook' && (
              <div className="mb-6 p-4 bg-indigo-50 border border-indigo-100 rounded-xl text-indigo-900">
                <div className="flex items-center gap-2 mb-2">
                  <div className="w-2 h-2 rounded-full bg-indigo-500 animate-pulse" />
                  <span className="font-semibold text-sm uppercase tracking-wider">Webhook Endpoint</span>
                </div>
                <div className="flex items-center gap-2 bg-white/50 p-2 rounded-lg border border-indigo-200 font-mono text-xs">
                  <span className="px-1.5 py-0.5 bg-indigo-600 text-white rounded text-[10px] font-bold">POST</span>
                  <code className="break-all">
                    {typeof window !== 'undefined' ? window.location.origin.replace(':3000', ':5000') : ''}/api/webhooks/{id}
                  </code>
                </div>
                <p className="mt-2 text-xs text-indigo-600">Hit this URL to trigger execution remotely.</p>
              </div>
            )}

            <pre className="bg-gray-50 p-4 rounded-xl text-sm overflow-x-auto text-gray-800 border border-gray-200 font-mono">
              {JSON.stringify(workflow.definition, null, 2)}
            </pre>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="pb-4">
          <CardTitle className="text-lg">Workflow Versions</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <UITableHeader>
              <TableRow>
                <TableHead>Version</TableHead>
                <TableHead>Name</TableHead>
                <TableHead>Trigger</TableHead>
                <TableHead>Created</TableHead>
                <TableHead>Action</TableHead>
              </TableRow>
            </UITableHeader>
            <TableBody>
              {versions.map((v) => (
                <TableRow key={v.id}>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <span>v{v.version}</span>
                      {workflow.version === v.version && (
                        <Badge variant="success">Active</Badge>
                      )}
                    </div>
                  </TableCell>
                  <TableCell>{v.name}</TableCell>
                  <TableCell>{v.trigger_type}</TableCell>
                  <TableCell className="text-gray-500">
                    {new Date(v.created_at).toLocaleString()}
                  </TableCell>
                  <TableCell>
                    {canEdit && workflow.version !== v.version ? (
                      <Button
                        variant="secondary"
                        onClick={() => handleActivateVersion(v.version)}
                        disabled={activatingVersion === v.version}
                      >
                        {activatingVersion === v.version ? "Activating..." : "Set Active"}
                      </Button>
                    ) : (
                      <span className="text-gray-400">-</span>
                    )}
                  </TableCell>
                </TableRow>
              ))}
              {versions.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} className="text-center text-gray-500 py-8">
                    No versions yet
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="pb-4">
          <CardTitle className="text-lg">Recent Runs</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <UITableHeader>
              <TableRow>
                <TableHead>Run ID</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Started</TableHead>
              </TableRow>
            </UITableHeader>
            <TableBody>
              {runs.map((r) => (
                <TableRow key={r.id}>
                  <TableCell className="font-mono text-sm text-blue-600 hover:underline">
                    <Link href={`/dashboard/runs/${r.id}`}>
                      {r.id.split('-')[0]}
                    </Link>
                  </TableCell>
                  <TableCell>
                    <Badge variant={
                      r.status === 'success' ? 'success' :
                        r.status === 'failed' ? 'danger' :
                          'default'
                    }>
                      {r.status}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-gray-500">
                    {new Date(r.created_at).toLocaleString()}
                  </TableCell>
                </TableRow>
              ))}
              {runs.length === 0 && (
                <TableRow>
                  <TableCell colSpan={3} className="text-center text-gray-500 py-8">
                    No runs yet
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
