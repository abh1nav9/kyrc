import { Nav } from "./components/Nav";
import { Footer } from "./components/Footer";
import { OpenSourceBanner } from "./sections/OpenSourceBanner";
import { Hero } from "./sections/Hero";
import { Features } from "./sections/Features";
import { LeaderboardSection } from "./sections/LeaderboardSection";
import { Docs } from "./sections/Docs";
import { Identity } from "./sections/Identity";
import { Metrics } from "./sections/Metrics";
import { Architecture } from "./sections/Architecture";

// App is pure composition — each section lives in its own file under
// src/sections/, shared building blocks under src/components/, and the
// interactive widgets under src/widgets/.
export function App() {
  return (
    <div className="relative z-[1] mx-auto max-w-[1080px] overflow-x-hidden px-4 sm:px-6">
      <Nav />
      <OpenSourceBanner />
      <Hero />
      <Features />
      <LeaderboardSection />
      <Docs />
      <Identity />
      <Metrics />
      <Architecture />
      <Footer />
    </div>
  );
}
