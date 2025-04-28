"use client";

import { useParams } from "next/navigation";
import ThreadDetailPage from "@/app/forum/[threadId]/page";

// Renders the thread detail page for a specific chain
export default function ChainThreadDetailPage() {
  const params = useParams();
  const chainId = params.chainId as string;
  const threadId = params.threadId as string;
  return <ThreadDetailPage />;
}