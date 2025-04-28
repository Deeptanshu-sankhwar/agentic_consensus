import { API_CONFIG } from '@/config';

interface RegisterAgentParams {
    name: string;
    role: "producer" | "validator";
    traits: string[];
    style: string;
    influences: string[];
    mood: string;
}

interface RegisterAgentResponse {
    agentID: string;
    p2pPort: number;
    apiPort: number;
    message: string;
}

interface CreateChainParams {
    chain_id: string;
    genesis_prompt: string;
}

interface CreateChainResponse {
    message: string;
    chain_id: string;
    bootstrap_node: {
        p2p_port: number;
        api_port: number;
    };
}

export interface Chain {
    chain_id: string;
    name: string;
    agents: number;
    blocks: number;
}

export interface Validator {
    ID: string;
    Name: string;
    Traits: string[];
    Style: string;
    Influences: string[];
    Mood: string;
    CurrentPolicy: string;
}

interface Transaction {
    content: string;
    from: string;
    to: string;
    amount: number;
    fee: number;
    timestamp: number;
}

export class ApiError extends Error {
    constructor(
        message: string,
        public status?: number,
        public data?: any
    ) {
        super(message);
        this.name = 'ApiError';
    }
}

// Registers a new agent with the specified parameters and chain ID
export async function registerAgent(agent: RegisterAgentParams, chainId: string): Promise<RegisterAgentResponse> {
    try {
        const response = await fetch(`${API_CONFIG.BASE_URL}${API_CONFIG.ENDPOINTS.REGISTER_AGENT}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Chain-Id': chainId,
            },
            body: JSON.stringify({
                ...agent,
                endpoint: `${API_CONFIG.AGENT_SERVICE_URL}/${agent.role}`
            }),
        });

        const data = await response.json();

        if (!response.ok) {
            throw new ApiError(
                data.error || 'Failed to register agent',
                response.status,
                data
            );
        }

        return data as RegisterAgentResponse;
    } catch (error) {
        if (error instanceof ApiError) {
            throw error;
        }
        throw new ApiError(
            error instanceof Error ? error.message : 'Unknown error occurred'
        );
    }
}

// Creates a new chain with the specified parameters
export async function createChain(params: CreateChainParams): Promise<CreateChainResponse> {
    try {
        const response = await fetch(`${API_CONFIG.BASE_URL}${API_CONFIG.ENDPOINTS.CREATE_CHAIN}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(params),
        });

        const data = await response.json();

        if (!response.ok) {
            throw new ApiError(
                data.error || 'Failed to create chain',
                response.status,
                data
            );
        }

        return data as CreateChainResponse;
    } catch (error) {
        if (error instanceof ApiError) {
            throw error;
        }
        throw new ApiError(
            error instanceof Error ? error.message : 'Unknown error occurred'
        );
    }
}

// Retrieves a list of all available chains
export async function listChains(): Promise<Chain[]> {
    try {
        const response = await fetch(`${API_CONFIG.BASE_URL}${API_CONFIG.ENDPOINTS.FETCH_CHAINS}`);
        const data = await response.json();

        if (!response.ok) {
            throw new ApiError(
                data.error || 'Failed to fetch chains',
                response.status,
                data
            );
        }

        return data.chains;
    } catch (error) {
        if (error instanceof ApiError) {
            throw error;
        }
        throw new ApiError(
            error instanceof Error ? error.message : 'Unknown error occurred'
        );
    }
}

// Fetches all agents for a specific chain
export async function fetchAgents(chainId: string): Promise<Validator[]> {
    const response = await fetch(`${API_CONFIG.BASE_URL}${API_CONFIG.ENDPOINTS.FETCH_AGENTS}`, {
        headers: {
            'X-Chain-Id': chainId,
        },
    });
    const data = await response.json();
    if (!response.ok) {
        throw new ApiError(data.error || 'Failed to fetch validators');
    }
    
    const agents = data.agents || data;
    
    const result = agents.map((agent: any) => ({
        ID: agent.id || agent.ID,
        Name: agent.name || agent.Name,
        Traits: agent.traits || agent.Traits || [],
        Style: agent.style || agent.Style || '',
        Influences: agent.influences || agent.Influences || [],
        Mood: agent.mood || agent.Mood || '',
        CurrentPolicy: agent.currentPolicy || agent.CurrentPolicy || null
    }));
    
    return result;
}

// Proposes a new block for the specified chain
export async function proposeBlock(chainId: string): Promise<void> {
    const response = await fetch(`${API_CONFIG.BASE_URL}${API_CONFIG.ENDPOINTS.PROPOSE_BLOCK}?wait=true`, {
        method: 'POST',
        headers: {
            'X-Chain-Id': chainId,
        },
    });
    if (!response.ok) {
        throw new ApiError('Failed to propose block');
    }
}

// Submits a new transaction to the specified chain
export async function submitTransaction(transaction: Transaction, chainId: string): Promise<void> {
    transaction.content = "{\"title\":\"A possible novel approach to the Riemann Hypothesis (RH)\",\"abstract\":\"This paper analyzes the RH from the definition of the Riemann zeta function, trying to obtain possible links between the hypothesis and other generalized zeta functions. A possible path is discussed using the \u03b6 function expression involving the fractional part of x, with insights into the convergence region and its implications.\",\"content\":\"In this study, we explore a novel pathway to approach the Riemann Hypothesis using a reformulation of the \u03b6(s) function involving floor and fractional part integrals. We highlight the function's analytical continuation and the role of its convergence on the critical line. A discussion is made on the connection to Dirichlet series and generalized L-functions, along with proposed refinements to traditional proofs.\",\"author\":\"Vincenzo Mantova\",\"topic_tags\":[\"Riemann Hypothesis\",\"Analytic Number Theory\",\"Zeta Function\"],\"timestamp\":1710129999}"
    const response = await fetch(`${API_CONFIG.BASE_URL}${API_CONFIG.ENDPOINTS.SUBMIT_TRANSACTION}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'X-Chain-Id': chainId,
        },
        body: JSON.stringify(transaction),
    });

    if (!response.ok) {
        throw new ApiError('Failed to submit transaction');
    }

    const data = await response.json();
    return data;
}