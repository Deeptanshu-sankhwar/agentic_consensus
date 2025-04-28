"use client";

import { useParams } from 'next/navigation';
import AgentsPage from '@/app/agents/page';

// Renders the agents page for a specific chain
export default function ChainAgentsPage() {
  const params = useParams();
  const chainId = params.chainId as string;
  return <AgentsPage />;
}