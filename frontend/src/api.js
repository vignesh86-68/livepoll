const BASE_URL = '/api';

async function fetchAPI(endpoint, options = {}) {
  const url = `${BASE_URL}${endpoint}`;
  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    credentials: 'include', // Important for cookies
  });

  if (response.status === 204) {
    return null;
  }

  const data = await response.json().catch(() => null);

  if (!response.ok) {
    throw data || { error: 'An unknown error occurred' };
  }

  return data;
}

export const api = {
  // Auth
  signup: (data) => fetchAPI('/auth/signup', { method: 'POST', body: JSON.stringify(data) }),
  login: (data) => fetchAPI('/auth/login', { method: 'POST', body: JSON.stringify(data) }),
  logout: () => fetchAPI('/auth/logout', { method: 'POST' }),
  getMe: () => fetchAPI('/auth/me', { method: 'GET' }),

  // Polls
  createPoll: (data) => fetchAPI('/polls', { method: 'POST', body: JSON.stringify(data) }),
  getMyPolls: () => fetchAPI('/polls/mine', { method: 'GET' }),
  getPoll: (code) => fetchAPI(`/polls/${code}`, { method: 'GET' }),
  getPollResults: (code) => fetchAPI(`/polls/${code}/results`, { method: 'GET' }),
  vote: (code, optionIds) => fetchAPI(`/polls/${code}/vote`, { method: 'POST', body: JSON.stringify({ optionIds }) }),
  closePoll: (code) => fetchAPI(`/polls/${code}/close`, { method: 'POST' }),
  deletePoll: (code) => fetchAPI(`/polls/${code}`, { method: 'DELETE' }),
};
