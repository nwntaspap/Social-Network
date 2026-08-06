'use client';

interface GroupSearchBarProps {
  query: string;
  onQueryChange: (query: string) => void;
}

export default function GroupSearchBar({ query, onQueryChange }: GroupSearchBarProps) {
  return (
    <div className="search-section">
      <input
        type="text"
        className="search-input"
        placeholder="Search groups..."
        value={query}
        onChange={(e) => onQueryChange(e.target.value)}
        aria-label="Search groups"
      />
    </div>
  );
}
