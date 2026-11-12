import {
  useEffect,
  useId,
  useRef,
  useState,
  type FormEvent,
  type MouseEvent,
} from "react";
import { ApiError, deleteLink, friendlyError, type AuthUser } from "../api/client";
import { navigate } from "../lib/navigation";
import { AlertIcon, CheckIcon, CloseIcon, LogOutIcon, UserIcon } from "./Icons";
import "./ManageDialog.css";

type Props = {
  open: boolean;
  token: string;
  user: AuthUser | null;
  isAuthenticated: boolean;
  onClose: () => void;
  onSignIn: () => void;
  onSessionExpired: () => void;
  onDeleted: (shortCode: string) => void;
};

type DeleteStatus = "idle" | "loading" | "success" | "error";

export function ManageDialog({
  open,
  token,
  user,
  isAuthenticated,
  onClose,
  onSignIn,
  onSessionExpired,
  onDeleted,
}: Props) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const deleteInputRef = useRef<HTMLInputElement>(null);
  const deleteStartRef = useRef<HTMLButtonElement>(null);
  const confirmButtonRef = useRef<HTMLButtonElement>(null);
  const signInRef = useRef<HTMLButtonElement>(null);

  const [deleteCode, setDeleteCode] = useState("");
  const [deleteStatus, setDeleteStatus] = useState<DeleteStatus>("idle");
  const [deleteMessage, setDeleteMessage] = useState<string | null>(null);
  const [confirming, setConfirming] = useState(false);

  const titleId = useId();
  const descriptionId = useId();
  const codeHintId = useId();

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) {
      return;
    }

    if (open && !dialog.open) {
      dialog.showModal();
    } else if (!open && dialog.open) {
      dialog.close();
    }
  }, [open]);

  useEffect(() => {
    if (!open) {
      return;
    }

    setDeleteCode("");
    setDeleteStatus("idle");
    setDeleteMessage(null);
    setConfirming(false);
  }, [open]);

  useEffect(() => {
    if (!open) {
      return;
    }

    const frame = window.requestAnimationFrame(() => {
      if (isAuthenticated) {
        deleteInputRef.current?.focus();
      } else {
        signInRef.current?.focus();
      }
    });
    return () => window.cancelAnimationFrame(frame);
  }, [open, isAuthenticated]);

  useEffect(() => {
    if (confirming) {
      confirmButtonRef.current?.focus();
    }
  }, [confirming]);

  function handleBackdropClick(event: MouseEvent<HTMLDialogElement>) {
    if (event.target === event.currentTarget) {
      onClose();
    }
  }

  function handleDeleteCodeChange(value: string) {
    setDeleteCode(value);
    setConfirming(false);
    if (deleteStatus !== "idle") {
      setDeleteStatus("idle");
      setDeleteMessage(null);
    }
  }

  function handleDeleteStart(event: FormEvent) {
    event.preventDefault();
    const code = deleteCode.trim();

    setDeleteMessage(null);
    if (!isAuthenticated || !token.trim()) {
      setDeleteStatus("error");
      setDeleteMessage("Sign in before deleting a link.");
      signInRef.current?.focus();
      return;
    }

    if (!code) {
      setDeleteStatus("error");
      setDeleteMessage("Enter the short code you want to delete.");
      deleteInputRef.current?.focus();
      return;
    }

    setDeleteStatus("idle");
    setConfirming(true);
  }

  function handleDeleteConfirmStart() {
    setConfirming(false);
    window.requestAnimationFrame(() => deleteStartRef.current?.focus());
  }

  async function handleDeleteConfirm() {
    const code = deleteCode.trim();
    setDeleteStatus("loading");
    setDeleteMessage(null);

    try {
      await deleteLink(code, token.trim());
      setDeleteStatus("success");
      setDeleteMessage(`Deleted “${code}”. The short link now returns 404.`);
      setDeleteCode("");
      setConfirming(false);
      onDeleted(code);
      window.requestAnimationFrame(() => deleteStartRef.current?.focus());
    } catch (error) {
      const expired = error instanceof ApiError && error.status === 401;
      setDeleteStatus("error");
      setDeleteMessage(
        expired
          ? "Your session expired. Sign in again to delete links."
          : friendlyError(error),
      );
      setConfirming(false);
      if (expired) {
        onSessionExpired();
        window.requestAnimationFrame(() => signInRef.current?.focus());
      } else {
        window.requestAnimationFrame(() => deleteStartRef.current?.focus());
      }
    }
  }

  function handleSignOut() {
    onClose();
    navigate("/signout");
  }

  return (
    <dialog
      ref={dialogRef}
      className="manage-dialog"
      aria-labelledby={titleId}
      aria-describedby={descriptionId}
      onClose={onClose}
      onClick={handleBackdropClick}
    >
      <div className="manage-dialog__inner">
        <header className="manage-dialog__head">
          <div>
            <h2 id={titleId} className="manage-dialog__title">
              Manage links
            </h2>
            <p id={descriptionId} className="manage-dialog__lede">
              Delete a link you own. Ownership is tied to the account that created it.
            </p>
          </div>
          <button
            className="manage-dialog__close"
            type="button"
            onClick={onClose}
            aria-label="Close manage links"
          >
            <CloseIcon className="icon" />
          </button>
        </header>

        <section className="manage-account" aria-label="Account">
          {isAuthenticated && user ? (
            <>
              <div className="person manage-account__person">
                <span className="person__avatar" aria-hidden="true">
                  {(user.display_name || user.email).charAt(0).toUpperCase()}
                </span>
                <span className="person__body">
                  <span className="person__name">{user.display_name || user.email}</span>
                  <span className="person__email">{user.email}</span>
                </span>
                <span className="manage-account__status">
                  <span className="manage-account__dot" aria-hidden="true" />
                  Signed in
                </span>
              </div>
              <button
                className="btn btn--quiet manage-account__signout"
                type="button"
                onClick={handleSignOut}
              >
                <LogOutIcon className="icon" />
                Sign out
              </button>
            </>
          ) : (
            <>
              <div className="person manage-account__person">
                <span className="person__avatar is-muted" aria-hidden="true">
                  <UserIcon className="icon" />
                </span>
                <span className="person__body">
                  <span className="person__name">Sign in to manage links</span>
                  <span className="person__email">
                    Links you create while signed in can be deleted here.
                  </span>
                </span>
              </div>
              <button
                ref={signInRef}
                className="btn btn--primary manage-account__signin"
                type="button"
                onClick={onSignIn}
              >
                Sign in
              </button>
            </>
          )}
        </section>

        <section className="manage-dialog__section" aria-labelledby="manage-delete-title">
          <h3 id="manage-delete-title" className="manage-dialog__section-title">
            Delete a link
          </h3>
          <p className="manage-dialog__section-copy">
            Deleting is permanent. The short code stops redirecting right away.
          </p>

          <form className="manage-delete" onSubmit={handleDeleteStart}>
            <label className="field" htmlFor="delete-code">
              <span className="field__label">Short code</span>
              <input
                id="delete-code"
                ref={deleteInputRef}
                className="field__control manage-delete__input"
                type="text"
                name="short_code"
                placeholder="my-link"
                maxLength={16}
                value={deleteCode}
                onChange={(event) => handleDeleteCodeChange(event.target.value)}
                autoComplete="off"
                autoCapitalize="none"
                spellCheck={false}
                disabled={!isAuthenticated}
                aria-describedby={codeHintId}
              />
              <span id={codeHintId} className="field__hint">
                {isAuthenticated
                  ? "Use the code at the end of the short link."
                  : "Sign in above to delete links."}
              </span>
            </label>

            {confirming ? (
              <div className="manage-confirm" role="group" aria-label="Confirm deletion">
                <p>
                  Delete <code>{deleteCode.trim()}</code>? This cannot be undone.
                </p>
                <div className="manage-confirm__actions">
                  <button
                    className="btn btn--danger"
                    type="button"
                    ref={confirmButtonRef}
                    onClick={handleDeleteConfirm}
                    disabled={deleteStatus === "loading"}
                  >
                    {deleteStatus === "loading" ? (
                      <>
                        <span className="spinner" aria-hidden="true" />
                        Deleting…
                      </>
                    ) : (
                      "Delete permanently"
                    )}
                  </button>
                  <button
                    className="btn btn--ghost"
                    type="button"
                    onClick={handleDeleteConfirmStart}
                    disabled={deleteStatus === "loading"}
                  >
                    Keep link
                  </button>
                </div>
              </div>
            ) : (
              <button
                ref={deleteStartRef}
                className="btn btn--danger-ghost manage-delete__submit"
                type="submit"
                disabled={!isAuthenticated || !token.trim() || deleteStatus === "loading"}
              >
                Delete link
              </button>
            )}

            {deleteMessage ? (
              <p
                className={deleteStatus === "success" ? "form-success" : "form-error"}
                role={deleteStatus === "error" ? "alert" : "status"}
              >
                {deleteStatus === "success" ? (
                  <CheckIcon className="icon" />
                ) : (
                  <AlertIcon className="icon" />
                )}
                <span>{deleteMessage}</span>
              </p>
            ) : null}
          </form>
        </section>
      </div>
    </dialog>
  );
}
