import Feed from './Feed';
import SearchBar from './SearchBar';
import SuggestedUsers from './SuggestedUsers';

export default function HomeContent() {
  return (
    <div className="home-container">
      <SearchBar />
      <SuggestedUsers />
      <Feed />
    </div>
  );
}
