"use client";

import { useState, useEffect } from "react";
import { Line } from "react-chartjs-2";
import {
  Chart as ChartJS,
  LineElement,
  CategoryScale,
  LinearScale,
  TimeScale,
  PointElement,
  Filler,
  Tooltip,
  Legend,
} from "chart.js";
import "chartjs-adapter-date-fns"; // Date adapter for time scale

// Register required Chart.js components
ChartJS.register(
  LineElement,
  CategoryScale,
  LinearScale,
  TimeScale,
  PointElement,
  Filler,
  Tooltip,
  Legend
);

// Define the type for each log entry
interface LogEntry {
  timestamp: string; // ISO 8601 string
  blood_sugar: number; // Blood sugar value
}

// Define the type for chart data
interface ChartData {
  labels: string[]; // Timestamps
  datasets: {
    label: string;
    data: number[];
    borderColor: string;
    backgroundColor: string;
    fill: boolean;
  }[];
}

export default function BloodSugarGraph() {
  const [chartData, setChartData] = useState<ChartData>({
    labels: [], // Timestamps
    datasets: [
      {
        label: "Blood Sugar Levels",
        data: [], // Blood sugar values
        borderColor: "rgba(75,192,192,1)",
        backgroundColor: "rgba(75,192,192,0.2)",
        fill: true,
      },
    ],
  });

  useEffect(() => {
    const fetchData = async () => {
      try {
        const response = await fetch(
          process.env.NEXT_PUBLIC_BASE_API_URL + "/logs"
        );
        const data: LogEntry[] = await response.json(); // Explicitly typed

        // Extract timestamps and blood sugar values
        const timestamps = data.map((entry) => entry.timestamp);
        const bloodSugarValues = data.map((entry) => entry.blood_sugar);

        setChartData({
          labels: timestamps,
          datasets: [
            {
              label: "Blood Sugar Levels",
              data: bloodSugarValues,
              borderColor: "rgba(75,192,192,1)",
              backgroundColor: "rgba(75,192,192,0.2)",
              fill: true,
            },
          ],
        });
      } catch (error) {
        console.error("Error fetching data:", error);
      }
    };

    fetchData();
  }, []);

  return (
    <div>
      <h2>Blood Sugar Levels</h2>
      <Line
        data={chartData}
        options={{
          responsive: true,
          scales: {
            x: {
              type: "time", // Use the registered 'time' scale
              time: {
                unit: "minute",
              },
            },
            y: {
              beginAtZero: true,
            },
          },
        }}
      />
    </div>
  );
}
