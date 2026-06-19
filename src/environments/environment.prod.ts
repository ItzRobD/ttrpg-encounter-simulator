export const environment = {
  name: 'production',
  // Relative path: works when the API is served behind the same origin as the
  // SPA (reverse proxy). Override with an absolute URL if the API lives on a
  // different host, e.g. 'https://api.your-domain.com/api/v1'.
  apiUrl: '/api/v1',
  httpRetryCount: 3,
  httpRetryDelay: 1000,
  limits: {
    maxCharacters: 8,
    maxMonsters: 15,
    maxTotal: 23
  },
  layout: {
    combatantPanelMaxWidth: '70%',
    combatantPanelCollapsedWidth: '16rem',
    breakpointXL: '1200px'
  },
  features: {
    gibbering: false,
    consumeLife: false
  },
  config: {
    disableIncludeLogsToggle: true,
    defaultIncludeLogs: true,
    defaultNumberOfRuns: 20,
    defaultMaxRounds: 30,
  }
};
