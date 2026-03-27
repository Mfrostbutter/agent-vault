import { useState, useEffect } from "react";
import { StatusBadge, LoadingSpinner, ErrorBanner, timeAgo, timeUntil } from "../../components/shared";
import DataTable, { type Column } from "../../components/DataTable";
import DropdownMenu from "../../components/DropdownMenu";

interface AgentRow {
  name: string;
  vault_id: string;
  vault_name: string;
  status: string;
  created_at: string;
  session_expires_at?: string;
}

function RowActions({
  agent,
  onDone,
}: {
  agent: AgentRow;
  onDone: () => void;
}) {
  if (agent.status === "revoked") return null;

  async function handleRevoke() {
    await fetch(
      `/v1/admin/agents/${encodeURIComponent(agent.name)}`,
      { method: "DELETE" }
    );
    onDone();
  }

  return (
    <DropdownMenu
      width={192}
      items={[
        { label: "Revoke agent", onClick: handleRevoke, variant: "danger" },
      ]}
    />
  );
}

export default function InstanceAgentsTab() {
  const [rows, setRows] = useState<AgentRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const columns: Column<AgentRow>[] = [
    {
      key: "name",
      header: "Name",
      render: (agent) => (
        <span className="text-sm font-mono font-medium text-text">
          {agent.name}
        </span>
      ),
    },
    {
      key: "vault",
      header: "Vault",
      render: (agent) => (
        <span className="text-sm text-text-muted">
          {agent.vault_name || agent.vault_id}
        </span>
      ),
    },
    {
      key: "status",
      header: "Status",
      render: (agent) => <StatusBadge status={agent.status} />,
    },
    {
      key: "created",
      header: "Last Seen",
      render: (agent) => (
        <span className="text-sm text-text-muted">
          {timeAgo(agent.created_at)}
        </span>
      ),
    },
    {
      key: "session_expires",
      header: "Session Expires",
      render: (agent) => {
        if (!agent.session_expires_at) {
          return <span className="text-sm text-text-dim">{"\u2014"}</span>;
        }
        const label = timeUntil(agent.session_expires_at);
        const isExpired = label === "Expired";
        return (
          <span className={`text-sm ${isExpired ? "text-danger" : "text-text-muted"}`}>
            {label}
          </span>
        );
      },
    },
    {
      key: "actions",
      header: "",
      align: "right" as const,
      render: (agent: AgentRow) => (
        <RowActions agent={agent} onDone={fetchData} />
      ),
    },
  ];

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 5000);
    return () => clearInterval(interval);
  }, []);

  async function fetchData() {
    try {
      const [agentsResp, vaultsResp] = await Promise.all([
        fetch("/v1/admin/agents"),
        fetch("/v1/admin/vaults"),
      ]);

      if (!agentsResp.ok) {
        const data = await agentsResp.json();
        setError(data.error || "Failed to load agents.");
        return;
      }

      // Build vault ID → name map
      const vaultMap: Record<string, string> = {};
      if (vaultsResp.ok) {
        const vaultsData = await vaultsResp.json();
        for (const v of vaultsData.vaults ?? []) {
          vaultMap[v.id] = v.name;
        }
      }

      const agentsData = await agentsResp.json();
      const agentRows: AgentRow[] = (agentsData.agents ?? []).map(
        (a: { name: string; vault_id: string; status: string; created_at: string; session_expires_at?: string }) => ({
          name: a.name,
          vault_id: a.vault_id,
          vault_name: vaultMap[a.vault_id] || a.vault_id,
          status: a.status,
          created_at: a.created_at,
          session_expires_at: a.session_expires_at,
        })
      );

      setRows(agentRows);
    } catch {
      setError("Network error.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="p-8 w-full max-w-[960px]">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-[22px] font-semibold text-text tracking-tight mb-1">
            Agents
          </h2>
          <p className="text-sm text-text-muted">
            All persistent agents across the instance.
          </p>
        </div>
      </div>

      {loading ? (
        <LoadingSpinner />
      ) : error ? (
        <ErrorBanner message={error} />
      ) : (
        <DataTable
          columns={columns}
          data={rows}
          rowKey={(row) => row.name}
          emptyTitle="No agents"
          emptyDescription="No persistent agents have been registered yet."
        />
      )}
    </div>
  );
}
