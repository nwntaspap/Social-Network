'use client';

import { useCallback, useEffect, useState } from 'react';
import { browseGroups } from '@/lib/api';
import type { Group, MembershipStatus } from '@/lib/types';
import GroupCard from './GroupCard';

export default function GroupList() {
  const [groups, setGroups] = useState<Group[]>([]);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState('');

  const loadPage = useCallback(async (nextPage: number) => {
    try {
      const response = await browseGroups('', nextPage);
      setGroups((prev) => (nextPage === 1 ? response.data : [...prev, ...response.data]));
      setTotalPages(response.totalPages || 1);
    } catch {
      setError('Failed to load groups. Please try again.');
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  }, []);

  useEffect(() => {
    let ignore = false;
    browseGroups('', 1)
      .then((response) => {
        if (ignore) return;
        setGroups(response.data);
        setTotalPages(response.totalPages || 1);
      })
      .catch(() => {
        if (!ignore) setError('Failed to load groups. Please try again.');
      })
      .finally(() => {
        if (!ignore) setLoading(false);
      });
    return () => {
      ignore = true;
    };
  }, []);

  const handleStatusChange = (groupId: string, newStatus: MembershipStatus) => {
    setGroups((prevGroups) =>
      prevGroups.map((group) =>
        group.id === groupId ? { ...group, membershipStatus: newStatus } : group
      )
    );
  };

  if (loading) {
    return <div className="groups-grid-loading">Loading groups...</div>;
  }

  if (error) {
    return <div className="groups-grid-error">{error}</div>;
  }

  return (
    <div>
      <div className="groups-grid">
        {groups.length > 0 ? (
          groups.map((group) => (
            <GroupCard key={group.id} group={group} onStatusChange={handleStatusChange} />
          ))
        ) : (
          <p className="groups-empty">No groups found.</p>
        )}
      </div>

      {page < totalPages && (
        <div className="feed-load-more">
          <button
            className="feed-load-more-btn"
            disabled={loadingMore}
            onClick={async () => {
              const nextPage = page + 1;
              setLoadingMore(true);
              setPage(nextPage);
              await loadPage(nextPage);
            }}
          >
            {loadingMore ? 'Loading...' : 'Load more'}
          </button>
        </div>
      )}
    </div>
  );
}
