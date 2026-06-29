import SearchBar from '@/components/features/home/SearchBar';
import SuggestedUsers from '@/components/features/home/SuggestedUsers';
import Feed from '@/components/features/home/Feed';

export default function HomeContent() {
  return (
    <div className="home-container">
      <SearchBar />
      <SuggestedUsers />
      <Feed />
    </div>
  );
}
