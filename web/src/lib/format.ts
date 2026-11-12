const relativeTime = new Intl.RelativeTimeFormat(undefined, { numeric: "auto" });
const fullDateTime = new Intl.DateTimeFormat(undefined, {
  dateStyle: "medium",
  timeStyle: "short",
});
const shortDate = new Intl.DateTimeFormat(undefined, {
  month: "short",
  day: "numeric",
});

export function formatDateTime(value: string): string {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : fullDateTime.format(date);
}

export function formatShortDate(value: string): string {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : shortDate.format(date);
}

export function formatRelativeTime(value: string, now = Date.now()): string {
  const date = new Date(value);
  const time = date.getTime();

  if (Number.isNaN(time)) {
    return "";
  }

  const diff = Math.round((time - now) / 1000);
  const abs = Math.abs(diff);

  if (abs < 45) {
    return "Just now";
  }
  if (abs < 3600) {
    return relativeTime.format(Math.round(diff / 60), "minute");
  }
  if (abs < 86_400) {
    return relativeTime.format(Math.round(diff / 3600), "hour");
  }
  if (abs < 2_592_000) {
    return relativeTime.format(Math.round(diff / 86_400), "day");
  }

  return shortDate.format(date);
}

export function isExpired(value?: string): boolean {
  if (!value) {
    return false;
  }

  const time = new Date(value).getTime();
  return !Number.isNaN(time) && time <= Date.now();
}
