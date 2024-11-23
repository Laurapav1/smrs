"use client";

import { useState, useEffect } from "react";
import axios from "axios";
import styles from "./page.module.css";

// Define the type of the API response
interface ApiResponse {
  level: number;
  state: string;
}

// Map states to background colors
const stateColors: Record<string, string> = {
  low: "#FFFF00", // Yellow
  normal: "#00FF00", // Green
  high: "#FFA500", // Orange
  critical: "#FF0000", // Red
};

export default function Home() {
  const [bloodSugar, setBloodSugar] = useState<number | null>(null);
  const [state, setState] = useState<string>("");

  useEffect(() => {
    // Create a CancelToken to cancel the request if the component unmounts
    const source = axios.CancelToken.source();

    const fetchData = async () => {
      try {
        const response = await axios.get<ApiResponse>(
          `${process.env.NEXT_PUBLIC_BASE_API_URL}/insulin-alarm`,
          {
            cancelToken: source.token,
            headers: { "Cache-Control": "no-cache" },
          }
        );
        setBloodSugar(response.data.level);
        setState(response.data.state);
      } catch (error) {
        if (axios.isCancel(error)) {
          console.log("Request canceled:", error.message);
        } else {
          console.error("Error fetching data:", error);
        }
      }
    };

    // Fetch data immediately and set an interval to fetch every second
    fetchData();
    const interval = setInterval(fetchData, 1000);

    // Cleanup: clear interval and cancel the request on component unmount
    return () => {
      clearInterval(interval);
      source.cancel("Component unmounted, request canceled");
    };
  }, []);

  return (
    <div
      className={styles.page}
      style={{
        backgroundColor: stateColors[state] || "#FFFFFF", // Default to white
      }}
    >
      <h1 className={styles.title}>Blood Sugar Monitor</h1>
      <p className={styles.value}>
        Blood sugar level: {bloodSugar !== null ? bloodSugar : "Loading..."}
      </p>
      <p className={styles.state}>State: {state || "Loading..."}</p>
    </div>
  );
}
