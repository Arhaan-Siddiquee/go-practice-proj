import { useState } from 'react';
import axios from 'axios';
import { FiGithub, FiSearch, FiLoader } from 'react-icons/fi';

export default function Home() {
  const [username, setUsername] = useState('');
  const [roast, setRoast] = useState('');
  const [stats, setStats] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  const handleRoast = async () => {
    if (!username) {
      setError('Please enter a GitHub username');
      return;
    }

    setIsLoading(true);
    setError('');
    setRoast('');
    setStats(null);
    
    try {
      const response = await axios.get(`http://localhost:8080/roast?username=${username}`);
      setRoast(response.data.roast);
      setStats(response.data.stats);
    } catch (err) {
      if (err.response?.status === 429) {
        setError('GitHub API rate limit exceeded. Try again later or add a token.');
      } else if (err.response?.status === 404) {
        setError('GitHub user not found. Check the username.');
      } else {
        setError(err.response?.data?.error || 'Failed to fetch roast. Please try again.');
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-3xl mx-auto">
        <div className="text-center mb-8">
          <h1 className="text-4xl font-extrabold text-gray-900 mb-2">
            GitHub Commit Roaster
          </h1>
          <p className="text-xl text-gray-600">
            We'll analyze your recent commits and roast you accordingly
          </p>
        </div>

        <div className="bg-white shadow rounded-lg p-6 mb-8">
          <div className="flex">
            <div className="relative flex-grow">
              <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <FiGithub className="h-5 w-5 text-gray-400" />
              </div>
              <input
                type="text"
                className="block w-full pl-10 pr-12 py-3 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                placeholder="GitHub username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && handleRoast()}
              />
              <div className="absolute inset-y-0 right-0 flex items-center">
                <button
                  onClick={handleRoast}
                  disabled={isLoading}
                  className="px-4 py-2 bg-indigo-600 text-white rounded-r-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 disabled:opacity-50 flex items-center"
                >
                  {isLoading ? (
                    <>
                      <FiLoader className="animate-spin mr-2" />
                      Roasting...
                    </>
                  ) : (
                    <>
                      <FiSearch className="mr-2" />
                      Roast Me
                    </>
                  )}
                </button>
              </div>
            </div>
          </div>
          {error && (
            <p className="mt-2 text-sm text-red-600">
              {error}
              {error.includes('rate limit') && (
                <span className="block mt-1">
                  Add a GITHUB_TOKEN to server/.env for higher limits
                </span>
              )}
            </p>
          )}
        </div>

        {stats && (
          <div className="bg-white shadow rounded-lg p-6 mb-8">
            <h2 className="text-xl font-semibold text-gray-900 mb-4">
              Stats for {username}
            </h2>
            <div className="grid grid-cols-2 gap-4">
              <div className="bg-gray-50 p-4 rounded-lg">
                <p className="text-sm font-medium text-gray-500">Total Commits</p>
                <p className="text-2xl font-semibold text-gray-900">
                  {stats.total_commits}
                </p>
              </div>
              <div className="bg-gray-50 p-4 rounded-lg">
                <p className="text-sm font-medium text-gray-500">Repos Analyzed</p>
                <p className="text-2xl font-semibold text-gray-900">
                  {stats.repos_analyzed}
                </p>
              </div>
            </div>
          </div>
        )}

        {roast && (
          <div className="bg-white shadow rounded-lg p-6">
            <h2 className="text-xl font-semibold text-gray-900 mb-4">Your Roast</h2>
            <div className="prose prose-indigo max-w-none">
              {roast.split('\n\n').map((paragraph, i) => (
                <p key={i} className="mb-4 text-gray-800">
                  {paragraph}
                </p>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}