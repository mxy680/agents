import Link from "next/link";

export default function Home() {
  return (
    <div className="flex min-h-[calc(100vh-4rem)] items-center justify-center">
      <div className="max-w-2xl px-4 text-center">
        <h1 className="mb-4 text-4xl font-bold tracking-tight text-gray-900 sm:text-5xl">
          Agent Marketplace
        </h1>
        <p className="mb-8 text-lg text-gray-600">
          Connect your Google, GitHub, and Instagram accounts so AI agents can
          work on your behalf. Tokens are encrypted at rest and never exposed
          directly.
        </p>
        <Link
          href="/integrations"
          className="inline-block rounded-md bg-gray-900 px-6 py-3 text-base font-medium text-white hover:bg-gray-800 transition-colors"
        >
          Get Started
        </Link>
      </div>
    </div>
  );
}
