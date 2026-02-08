export const environment = {
  name: 'production',
  apiUrl: 'http://localhost:8080/api/v1',
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
