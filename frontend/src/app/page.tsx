"use client";

import { useState, useEffect } from "react";
import axios from "axios";
import styles from "./page.module.css";

// Define the type of the API response
interface ApiResponse {
  message: string;
}

export default function Home() {
  const [bloodSugar, setBloodSugar] = useState<string>("");

  console.log(process.env);
  useEffect(() => {
    console.log(process.env.NEXT_PUBLIC_BASE_API_URL);
    // Create a CancelToken to cancel the request if the component unmounts
    const source = axios.CancelToken.source();
    console.log(process.env);

    const fetchData = async () => {
      try {
        const response = await axios.get<ApiResponse>(
          process.env.NEXT_PUBLIC_BASE_API_URL + "/insulin-alarm",
          {
            cancelToken: source.token,
            headers: { "Cache-Control": "no-cache" },
          }
        );
        setBloodSugar(response.data.message);
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
    const interval = setInterval(fetchData, 1000000);

    // Cleanup: clear interval and cancel the request on component unmount
    return () => {
      clearInterval(interval);
      source.cancel("Component unmounted, request canceled");
    };
  }, []);

  return (
    <div className={styles.page}>
      <p>jkgkjlsfg{bloodSugar}</p>
    </div>
  );
}
