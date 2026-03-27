import { useState, useEffect } from "react";
import { LoadingSpinner, ErrorBanner, timeAgo } from "../../components/shared";
import DataTable, { type Column } from "../../components/DataTable";

interface InstanceVault {
  id: string;
  name: string;
  created_at: string;
}

export default function InstanceVaultsTab() {
  const [vaults, setVaults] = useState<InstanceVault[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    fetchVaults();
  }, []);

  async function fetchVaults() {
    try {
      const resp = await fetch("/v1/admin/vaults");
      if (!resp.ok) {
        const data = await resp.json();
        setError(data.error || "Failed to load vaults.");
        return;
      }
      const data = await resp.json();
      setVaults(data.vaults ?? []);
    } catch {
      setError("Network error.");
    } finally {
      setLoading(false);
    }
  }

  const columns: Column<InstanceVault>[] = [
    {
      key: "name",
      header: "Name",
      render: (v) => <span className="text-sm text-text font-medium">{v.name}</span>,
    },
    {
      key: "created_at",
      header: "Created",
      render: (v) => (
        <span className="text-sm text-text-muted">{timeAgo(v.created_at)}</span>
      ),
    },
  ];

  return (
    <div className="p-8 w-full max-w-[960px]">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-[22px] font-semibold text-text tracking-tight mb-1">
            Vaults
          </h2>
          <p className="text-sm text-text-muted">
            All vaults across the instance.
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
          data={vaults}
          rowKey={(v) => v.id}
          emptyTitle="No vaults"
          emptyDescription="No vaults have been created yet."
        />
      )}
    </div>
  );
}
