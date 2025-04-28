"use client";

import { useParams } from 'next/navigation';
import ForumPage from '@/app/forum/page';

// Renders the forum page for a specific chain
export default function ChainForumPage() {
  const params = useParams();
  const chainId = params.chainId as string;
  return <ForumPage />;
}