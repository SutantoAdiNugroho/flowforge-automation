"use client";

import { useEffect, useMemo, useState, use } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { fetchApi } from "@/lib/api";
import { useAuth } from "@/contexts/AuthContext";

import { Button } from "@/components/ui/Button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/Card";
import { Input } from "@/components/ui/Input";
import { Select } from "@/components/ui/Select";
import { Textarea } from "@/components/ui/Textarea";
import { Editor } from "@monaco-editor/react";

export default function NewWorkflowVersionPage({ params }: { params: Promise<{ id: string }> }) {
  const resolvedParams = use(params);
  const id = resolvedParams.id;
  const router = useRouter();
  const { hasRole } = useAuth();

  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const [mode, setMode] = useState<"import" | "new">("import");
  const [workflow, setWorkflow] = useState<any>(null);
  const [versions, setVersions] = useState<any[]>([]);
  const [importFromVersion, setImportFromVersion] = useState<number | null>(null);

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [triggerType, setTriggerType] = useState("manual");
  const [cronExpression, setCronExpression] = useState("");
  const [definition, setDefinition] = useState("{\n  \"steps\": []\n}");

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [wfData, versionsData] = await Promise.all([
          fetchApi(`/workflows/${id}`),
          fetchApi(`/workflows/${id}/versions?page=1&page_size=100`),
        ]);

        setWorkflow(wfData);
        setVersions(versionsData.content || []);

        const latestVersion = versionsData.content?.[0];
        if (latestVersion) {
          setImportFromVersion(latestVersion.version);
        }

        setName(wfData.name || "");
        setDescription(wfData.description || "");
        setTriggerType(wfData.trigger_type || "manual");
        setCronExpression(wfData.cron_expression || "");
        setDefinition(JSON.stringify(wfData.definition || { steps: [] }, null, 2));
      } catch (err: any) {
        setError(err.message || "Failed to load workflow data");
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [id]);

  const selectedImportVersion = useMemo(() => {
    if (!importFromVersion) return null;
    return versions.find((v) => v.version === importFromVersion) || null;
  }, [versions, importFromVersion]);

  useEffect(() => {
    if (mode === "import" && selectedImportVersion) {
      setName(selectedImportVersion.name || "");
      setDescription(selectedImportVersion.description || "");
      setTriggerType(selectedImportVersion.trigger_type || "manual");
      setCronExpression(selectedImportVersion.cron_expression || "");
      setDefinition(JSON.stringify(selectedImportVersion.definition || { steps: [] }, null, 2));
    } else if (mode === "new") {
      setName("");
      setDescription("");
      setTriggerType("manual");
      setCronExpression("");
      setDefinition(JSON.stringify({ steps: [] }, null, 2));
    }
  }, [mode, selectedImportVersion]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    let parsedDef;
    try {
      parsedDef = JSON.parse(definition);
    } catch {
      setError("DAG Definition harus JSON valid");
      return;
    }

    const body: any = {
      name,
      description,
      trigger_type: triggerType,
      cron_expression: triggerType === "cron" ? cronExpression : "",
      definition: parsedDef,
    };

    if (mode === "import" && importFromVersion) {
      body.import_from_version = importFromVersion;
    }

    try {
      setSaving(true);
      await fetchApi(`/workflows/${id}/versions`, {
        method: "POST",
        body: JSON.stringify(body),
      });
      router.refresh();
      router.push(`/dashboard/workflows/${id}`);
    } catch (err: any) {
      setError(err.message || "Failed to create new version");
    } finally {
      setSaving(false);
    }
  };

  if (!hasRole(["admin", "editor"])) {
    return <div className="p-4 text-red-600">Access Denied</div>;
  }

  if (loading) {
    return <div className="text-gray-500 text-center py-8">Loading...</div>;
  }

  return (
    <div className="max-w-3xl mx-auto py-8 px-4">
      <Card>
        <CardHeader>
          <CardTitle>New Workflow Version</CardTitle>
        </CardHeader>
        <CardContent>
          {error && <div className="mb-6 text-sm text-red-600 bg-red-50 p-4 rounded-lg border border-red-200">{error}</div>}

          <form onSubmit={handleSubmit} className="space-y-6">
            <Select
              label="Mode"
              value={mode}
              onChange={(e) => setMode(e.target.value as "import" | "new")}
            >
              <option value="import">import previous version</option>
              <option value="new">new</option>
            </Select>

            {mode === "import" && (
              <Select
                label="Source Version"
                value={importFromVersion?.toString() || ""}
                onChange={(e) => setImportFromVersion(Number(e.target.value))}
                required
              >
                <option value="" disabled>Select version to import</option>
                {versions.map((v) => (
                  <option key={v.id} value={v.version}>
                    v{v.version} - {v.name}
                  </option>
                ))}
              </Select>
            )}

            <div className="space-y-4 pt-4 border-t border-gray-100">
              <Input
                label="Name"
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
              />

              <Textarea
                label="Description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
              />

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Select
                  label="Trigger Type"
                  value={triggerType}
                  onChange={(e) => setTriggerType(e.target.value)}
                >
                  <option value="manual">Manual</option>
                  <option value="cron">Cron</option>
                  <option value="webhook">Webhook</option>
                </Select>

                {triggerType === "webhook" && (
                  <p className="mt-1 text-xs text-indigo-600 bg-indigo-50 p-2 rounded border border-indigo-100">
                    Webhook requires a POST request to the backend URL to trigger execution.
                  </p>
                )}

                {triggerType === "cron" && (
                  <Input
                    label="Cron Expression"
                    type="text"
                    value={cronExpression}
                    onChange={(e) => setCronExpression(e.target.value)}
                    placeholder="* * * * *"
                    required
                  />
                )}
              </div>

              <div className="space-y-2">
                <label className="block text-sm font-medium text-gray-700">DAG Definition (JSON)</label>
                <div className="border border-gray-200 rounded-md overflow-hidden bg-gray-50">
                  <Editor
                    height="400px"
                    defaultLanguage="json"
                    value={definition}
                    onChange={(value) => setDefinition(value || "")}
                    options={{
                      minimap: { enabled: false },
                      fontSize: 13,
                      scrollBeyondLastLine: false,
                      automaticLayout: true,
                      formatOnPaste: true,
                      formatOnType: true,
                    }}
                  />
                </div>
                <p className="text-xs text-gray-500">Hint: Press Shift+Alt+F to format JSON</p>
              </div>
            </div>

            <div className="flex justify-end space-x-3 pt-6 border-t border-gray-100">
              <Link href={`/dashboard/workflows/${id}`}>
                <Button type="button" variant="secondary">Cancel</Button>
              </Link>
              <Button type="submit" disabled={saving}>
                {saving ? "Creating..." : "Create Version"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
