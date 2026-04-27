const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:5000/api";

export async function fetchApi(endpoint: string, options: RequestInit = {}) {
  let token = "";
  if (typeof window !== "undefined") {
    token = localStorage.getItem("token") || "";
  }

  const headers = {
    "Content-Type": "application/json",
    ...(token && { Authorization: `Bearer ${token}` }),
    ...options.headers,
  };

  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers,
  });

  if (!response.ok) {
    if (response.status === 401 && typeof window !== "undefined") {
      localStorage.removeItem("token");
      localStorage.removeItem("user");
      window.location.href = "/login";
      return;
    }

    let errMessage = "API Error";
    try {
      const errData = await response.json();
      errMessage = errData.message || errData.error || errMessage;
    } catch (e) {}
    throw new Error(errMessage);
  }

  // Handle 204 No Content
  if (response.status === 204) return null;

  return response.json();
}
