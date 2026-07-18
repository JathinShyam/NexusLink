import { Hero } from "./components/Hero";
import { ManagePanel } from "./components/ManagePanel";
import { ResultPanel } from "./components/ResultPanel";
import { ShortenForm } from "./components/ShortenForm";
import { SiteFooter } from "./components/SiteFooter";
import { useAuthToken } from "./hooks/useAuthToken";
import { useShorten } from "./hooks/useShorten";
import "./App.css";

export default function App() {
  const { token, setToken, clearToken, hasToken } = useAuthToken();
  const { result, error, create, isLoading } = useShorten(token);

  return (
    <div className="app">
      <Hero />

      <main className="workspace">
        <div className="workspace__inner">
          <ShortenForm
            isLoading={isLoading}
            error={error}
            hasToken={hasToken}
            onSubmit={create}
          />
          {result ? <ResultPanel result={result} /> : null}
          <ManagePanel token={token} onTokenChange={setToken} onClearToken={clearToken} />
        </div>
      </main>

      <SiteFooter />
    </div>
  );
}
