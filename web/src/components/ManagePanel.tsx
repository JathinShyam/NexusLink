import { useState, type FormEvent } from "react";
import { deleteLink, friendlyError } from "../api/client";
import "./ManagePanel.css";

type Props = {
  token: string;
  onTokenChange: (value: string) => void;
  onClearToken: () => void;
};

export function ManagePanel({ token, onTokenChange, onClearToken }: Props) {
  const [shortCode, setShortCode] = useState("");
  const [status, setStatus] = useState<"idle" | "loading" | "success" | "error">("idle");
  const [message, setMessage] = useState<string | null>(null);

  async function handleDelete(e: FormEvent) {
    e.preventDefault();
    setStatus("loading");
    setMessage(null);

    if (!token.trim()) {
      setStatus("error");
      setMessage("Paste a JWT first (make gentoken).");
      return;
    }

    const code = shortCode.trim();
    if (!code) {
      setStatus("error");
      setMessage("Enter the short code to delete.");
      return;
    }

    try {
      await deleteLink(code, token.trim());
      setStatus("success");
      setMessage(`Deleted “${code}”. Redirects will return 404.`);
      setShortCode("");
    } catch (err) {
      setStatus("error");
      setMessage(friendlyError(err));
    }
  }

  return (
    <section id="manage" className="panel manage" aria-labelledby="manage-title">
      <div className="panel__head">
        <h2 id="manage-title" className="panel__title">
          Manage with JWT
        </h2>
        <p className="panel__lede">
          Run <code className="inline-code">make gentoken</code>, paste the token, create owned links
          above, then soft-delete by short code.
        </p>
      </div>

      <div className="manage__stack">
        <label className="field">
          <span className="field__label">Bearer token</span>
          <textarea
            className="field__textarea"
            name="jwt"
            rows={3}
            spellCheck={false}
            placeholder="eyJhbGciOiJIUzI1NiIs..."
            value={token}
            onChange={(e) => onTokenChange(e.target.value)}
          />
        </label>

        <div className="manage__token-actions">
          <button className="btn btn--ghost" type="button" onClick={onClearToken} disabled={!token}>
            Clear token
          </button>
          {token ? <span className="manage__badge">Token saved locally</span> : null}
        </div>

        <form className="manage__delete" onSubmit={handleDelete}>
          <label className="field">
            <span className="field__label">Short code to delete</span>
            <input
              className="field__input"
              type="text"
              name="short_code"
              placeholder="my-link"
              maxLength={16}
              value={shortCode}
              onChange={(e) => setShortCode(e.target.value)}
            />
          </label>

          {message ? (
            <p
              className={status === "success" ? "form-success" : "form-error"}
              role="status"
            >
              {message}
            </p>
          ) : null}

          <button
            className="btn btn--danger"
            type="submit"
            disabled={status === "loading"}
          >
            {status === "loading" ? "Deleting…" : "Delete link"}
          </button>
        </form>
      </div>
    </section>
  );
}
