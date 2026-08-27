import Feed from './Feed';
import SearchBar from './SearchBar';
import SuggestedUsers from './SuggestedUsers';

export default function HomeContent() {
  return (
    <div className="home-layout">
      <aside className="home-sidebar">
        <SuggestedUsers />
      </aside>
      <div className="home-main">
        <SearchBar />
        <Feed />
      </div>
    </div>
  );
}
