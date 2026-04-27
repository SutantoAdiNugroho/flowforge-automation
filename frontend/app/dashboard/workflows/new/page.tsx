"use client";

import { useState } from "react";
import { fetchApi } from "@/lib/api";
import { useRouter } from "next/navigation";
import { useAuth } from "@/contexts/AuthContext";
import Link from "next/link";

import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Select } from "@/components/ui/Select";
import { Textarea } from "@/components/ui/Textarea";
import { Editor } from "@monaco-editor/react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/Card";

export default function CreateWorkflowPage() {
  const router = useRouter();
  const { hasRole } = useAuth();
  
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [triggerType, setTriggerType] = useState("manual");
  const [cronExpression, setCronExpression] = useState("");
  const [definition, setDefinition] = useState("{\n  \"steps\": []\n}");
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);

  if (!hasRole(["admin", "editor"])) {
    return <div className="p-4 text-red-600">Access Denied</div>;
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    let parsedDef;
    try {
      parsedDef = JSON.parse(definition);
    } catch (err) {
      setError("Invalid JSON definition");
      return;
    }

    try {
      setSaving(true);
      await fetchApi("/workflows", {
        method: "POST",
        body: JSON.stringify({
          name,
          description,
          trigger_type: triggerType,
          cron_expression: triggerType === "cron" ? cronExpression : "",
          definition: parsedDef,
          is_active: true,
        }),
      });
      router.refresh();
      router.push("/dashboard");
    } catch (err: any) {
      setError(err.message || "Failed to create");
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="max-w-3xl mx-auto py-8 px-4">
      <Card>
        <CardHeader className="pb-4">
          <CardTitle>Create Workflow</CardTitle>
        </CardHeader>
        <CardContent className="p-6">
          {error && <div className="mb-6 text-sm text-red-600 bg-red-50 p-4 rounded-lg border border-red-200">{error}</div>}
          
          <form onSubmit={handleSubmit} className="space-y-6">
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
              <div className="space-y-1">
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
              </div>

              {triggerType === "cron" && (
                <Input
                  label="Cron Expression"
                  type="text"
                  placeholder="e.g. 0 9 * * *"
                  value={cronExpression}
                  onChange={(e) => setCronExpression(e.target.value)}
                  required
                />
              )}
            </div>
            
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <label className="block text-sm font-medium text-gray-700">DAG Definition (JSON)</label>
                <Link href="/dashboard/workflows/dag-guide">
                  <Button type="button" variant="secondary">DAG Guide</Button>
                </Link>
              </div>
              <div className="border border-gray-200 rounded-md overflow-hidden bg-gray-50">
                <Editor
                  height="300px"
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
            
            <div className="flex justify-end space-x-3 pt-6 border-t border-gray-100">
              <Link href="/dashboard">
                <Button type="button" variant="secondary">Cancel</Button>
              </Link>
              <Button type="submit" disabled={saving}>
                {saving ? "Creating..." : "Create Workflow"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
