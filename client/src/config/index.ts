// Configuration for API endpoints and HTTP status codes
export const API_CONFIG = {
    BASE_URL: 'http://127.0.0.1:3000',
    AGENT_SERVICE_URL: 'http://localhost:5000',
    ENDPOINTS: {
        REGISTER_AGENT: '/api/register',
        CREATE_CHAIN: '/api/chains',
        FETCH_CHAINS: '/api/chains',
        FETCH_AGENTS: '/api/agents',
        PROPOSE_BLOCK: '/api/block/propose',
        SUBMIT_TRANSACTION: '/api/transactions',
    },
    TIMEOUT: 10000,
} as const;

export const HTTP_STATUS = {
    OK: 200,
    BAD_REQUEST: 400,
    UNAUTHORIZED: 401,
    NOT_FOUND: 404,
    INTERNAL_SERVER_ERROR: 500,
} as const;