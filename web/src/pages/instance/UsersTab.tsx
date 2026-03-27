import { useState, useEffect } from "react";
import { LoadingSpinner, ErrorBanner, timeAgo } from "../../components/shared";
import DataTable, { type Column } from "../../components/DataTable";

interface InstanceUser {
  email: string;
  role: string;
  vaults: string[];
  created_at: string;
}

export default function InstanceUsersTab() {
  const [users, setUsers] = useState<InstanceUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    fetchUsers();
  }, []);

  async function fetchUsers() {
    try {
      const resp = await fetch("/v1/admin/users");
      if (!resp.ok) {
        const data = await resp.json();
        setError(data.error || "Failed to load users.");
        return;
      }
      const data = await resp.json();
      setUsers(data.users ?? []);
    } catch {
      setError("Network error.");
    } finally {
      setLoading(false);
    }
  }

  const columns: Column<InstanceUser>[] = [
    {
      key: "email",
      header: "Email",
      render: (u) => <span className="text-sm text-text">{u.email}</span>,
    },
    {
      key: "role",
      header: "Role",
      render: (u) => (
        <span className="text-sm text-text-muted capitalize">{u.role}</span>
      ),
    },
    {
      key: "vaults",
      header: "Vaults",
      render: (u) => (
        <span className="text-sm text-text-muted">
          {u.vaults && u.vaults.length > 0 ? u.vaults.join(", ") : "—"}
        </span>
      ),
    },
    {
      key: "created_at",
      header: "Created",
      render: (u) => (
        <span className="text-sm text-text-muted">{timeAgo(u.created_at)}</span>
      ),
    },
  ];

  return (
    <div className="p-8 w-full max-w-[960px]">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-[22px] font-semibold text-text tracking-tight mb-1">
            Users
          </h2>
          <p className="text-sm text-text-muted">
            All users across the instance.
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
          data={users}
          rowKey={(u) => u.email}
          emptyTitle="No users"
          emptyDescription="No users have registered yet."
        />
      )}
    </div>
  );
}
