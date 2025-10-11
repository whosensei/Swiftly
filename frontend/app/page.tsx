"use client"
import { Hero } from "@/components/hero";
import { Shorten } from "@/components/shorten";

export default function Home() {
  return (
    <div className="flex justify-center min-h-screen w-full overflow-x-hidden">
      <div className="flex flex-col min-h-screen w-full max-w-3xl border-l border-r border-dashed overflow-x-hidden">
        <div className="flex flex-col items-center justify-center flex-1 gap-8">
          <Hero />
          <Shorten />
        </div>
      </div>
    </div>
  );
}
