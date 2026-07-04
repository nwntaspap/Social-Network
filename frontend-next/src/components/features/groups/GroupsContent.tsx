import GroupSearchBar from './GroupSearchBar';
import GroupList from './GroupList';

export default function GroupsContent() {
  return (
    <div className="groups-container">
      <h1 className="groups-title">Groups</h1>
      <GroupSearchBar />
      <GroupList />
    </div>
  );
}
