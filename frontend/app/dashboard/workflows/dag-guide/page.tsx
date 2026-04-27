import Link from "next/link";

import { Button } from "@/components/ui/Button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/Card";

export default function DAGGuidePage() {
  return (
    <div className="max-w-5xl mx-auto py-8 px-4 space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>DAG Definition Guide</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-sm text-gray-700">
            use JSON object with root key <strong>steps</strong>. each step is one node in the workflow graph.
          </p>

          <div className="overflow-x-auto rounded-md border border-gray-200">
            <table className="min-w-full text-sm">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left font-semibold text-gray-700">Field</th>
                  <th className="px-4 py-3 text-left font-semibold text-gray-700">Function</th>
                  <th className="px-4 py-3 text-left font-semibold text-gray-700">Type</th>
                  <th className="px-4 py-3 text-left font-semibold text-gray-700">Required</th>
                  <th className="px-4 py-3 text-left font-semibold text-gray-700">Valid Value</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 bg-white">
                <tr>
                  <td className="px-4 py-3">steps</td>
                  <td className="px-4 py-3">List of execution nodes in this workflow.</td>
                  <td className="px-4 py-3">array</td>
                  <td className="px-4 py-3">Yes</td>
                  <td className="px-4 py-3">At least 1 step</td>
                </tr>
                <tr>
                  <td className="px-4 py-3">steps[].id</td>
                  <td className="px-4 py-3">Unique key used for dependency mapping.</td>
                  <td className="px-4 py-3">string</td>
                  <td className="px-4 py-3">Yes</td>
                  <td className="px-4 py-3">Unique, not empty, no self-duplicate</td>
                </tr>
                <tr>
                  <td className="px-4 py-3">steps[].name</td>
                  <td className="px-4 py-3">Readable label for monitoring and logs.</td>
                  <td className="px-4 py-3">string</td>
                  <td className="px-4 py-3">Yes</td>
                  <td className="px-4 py-3">Not empty</td>
                </tr>
                <tr>
                  <td className="px-4 py-3">steps[].type</td>
                  <td className="px-4 py-3">Chooses the runner behavior for this step.</td>
                  <td className="px-4 py-3">string</td>
                  <td className="px-4 py-3">Yes</td>
                  <td className="px-4 py-3">http, script, condition, parallel, delay</td>
                </tr>
                <tr>
                  <td className="px-4 py-3">steps[].config</td>
                  <td className="px-4 py-3">Step-specific settings consumed by the runner.</td>
                  <td className="px-4 py-3">object</td>
                  <td className="px-4 py-3">No</td>
                  <td className="px-4 py-3">Any JSON object by step type</td>
                </tr>
                <tr>
                  <td className="px-4 py-3">steps[].depends_on</td>
                  <td className="px-4 py-3">Defines execution order and dependency edges.</td>
                  <td className="px-4 py-3">array[string]</td>
                  <td className="px-4 py-3">No</td>
                  <td className="px-4 py-3">Only existing step IDs, no self reference, no cycle</td>
                </tr>
                <tr>
                  <td className="px-4 py-3">steps[].timeout</td>
                  <td className="px-4 py-3">Maximum execution duration per step.</td>
                  <td className="px-4 py-3">number (ms)</td>
                  <td className="px-4 py-3">No</td>
                  <td className="px-4 py-3">0 or &gt;= 1000</td>
                </tr>
                <tr>
                  <td className="px-4 py-3">steps[].retry_policy.max_retries</td>
                  <td className="px-4 py-3">Retry attempt limit after a failure.</td>
                  <td className="px-4 py-3">number</td>
                  <td className="px-4 py-3">No</td>
                  <td className="px-4 py-3">&gt;= 0</td>
                </tr>
                <tr>
                  <td className="px-4 py-3">steps[].retry_policy.initial_delay_ms</td>
                  <td className="px-4 py-3">First wait time before retry starts.</td>
                  <td className="px-4 py-3">number (ms)</td>
                  <td className="px-4 py-3">No</td>
                  <td className="px-4 py-3">&gt;= 100</td>
                </tr>
                <tr>
                  <td className="px-4 py-3">steps[].retry_policy.backoff_multiplier</td>
                  <td className="px-4 py-3">Scales delay between retry attempts.</td>
                  <td className="px-4 py-3">number</td>
                  <td className="px-4 py-3">No</td>
                  <td className="px-4 py-3">Recommended &gt; 1.0</td>
                </tr>
                <tr>
                  <td className="px-4 py-3">steps[].retry_policy.max_delay_ms</td>
                  <td className="px-4 py-3">Caps retry wait time growth.</td>
                  <td className="px-4 py-3">number (ms)</td>
                  <td className="px-4 py-3">No</td>
                  <td className="px-4 py-3">Recommended &gt;= initial_delay_ms</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div className="space-y-2">
            <h3 className="text-sm font-semibold text-gray-800">Minimal Valid Example</h3>
            <pre className="bg-gray-50 border border-gray-200 rounded-md p-4 text-xs overflow-x-auto text-gray-800">{`{
                "steps": [
                    {
                    "id": "step_1",
                    "name": "Get Data",
                    "type": "http",
                    "config": {
                        "method": "GET",
                        "url": "https://example.com/api"
                    },
                    "timeout": 5000
                    },
                    {
                    "id": "step_2",
                    "name": "Process Data",
                    "type": "script",
                    "depends_on": ["step_1"],
                    "retry_policy": {
                        "max_retries": 3,
                        "initial_delay_ms": 500,
                        "backoff_multiplier": 2,
                        "max_delay_ms": 10000
                    }
                    }
                ]
                }`}
            </pre>
          </div>

          <div className="flex justify-end">
            <Link href="/dashboard">
              <Button variant="secondary">Back</Button>
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
