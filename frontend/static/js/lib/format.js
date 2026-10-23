/**
 * Pure duration formatting shared by the SPA (no DOM).
 */

export function pad(num) {
  return String(num).padStart(2, '0');
}

/**
 * Converts seconds into a short human label: "0m", "45m", "1h 23m".
 */
export function formatSeconds(seconds) {
  if (!seconds || seconds <= 0) return '0m';
  const hrs = Math.floor(seconds / 3600);
  const mins = Math.floor((seconds % 3600) / 60);
  if (hrs > 0) return `${hrs}h ${pad(mins)}m`;
  return `${mins}m`;
}
