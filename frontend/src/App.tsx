import "./App.css";
import SearchForm from './components/SearchForm';
import Results from './components/Results';
import UploadForm from './components/UploadForm';
import StatsView from './components/StatsView';
import Navbar from './components/Navbar';

function App() {
  return (
    <div className="App">
      <Navbar/>
      <UploadForm />
      <SearchForm />
      <StatsView />
      <Results />
    </div>
  );
}

export default App;
