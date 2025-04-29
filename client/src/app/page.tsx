"use client";

import Link from "next/link";

// Main home page component with landing content and navigation
export default function Home() {
  return (
    <div className="w-full h-screen flex">
      <div className="w-full md:w-full bg-gray-950 flex flex-col h-full">
        <header className="p-8 pl-32 text-lg">
          <Link href="/" className="flex items-center gap-2">
            <span className="text-[#fd7653] font-bold">AGENTIC CONSENSUS</span>
          </Link>
        </header>

        <div className="px-32 p-8 flex flex-grow items-center">
          <div className="w-full flex flex-col">
            <div className="mb-8 text-gray-100">
              <h1 className="text-6xl font-bold mb-4">
                Redefining Governance in Blockchain
              </h1>
              <p className="text-gray-300">
                Explore a new frontier where AI meets decentralization. Create
                your own chain or collaborate with others as AI agents shape the
                evolution of governance.
              </p>
            </div>

            <div className="flex flex-col gap-4">
              <Link href="/agents">
                <button className="w-full bg-gradient-to-r from-[#fd7653] to-[#feb082] text-white font-medium font-semibold px-8 py-3 rounded-2xl hover:shadow-lg shadow-md transition-all duration-300 transform hover:-translate-y-0.5">
                  Launch Chain
                </button>
              </Link>
            </div>
          </div>
        </div>
      </div>

      <div className="w-full md:w-9/10 h-screen relative">
        <img
          src="/background_eth.png"
          alt="Agentic Consensus Logo"
          className="w-full h-full object-cover"
        />
      </div>
    </div>
  );
}
