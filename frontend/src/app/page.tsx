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
import "chartjs-adapter-date-fns";

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

// Define the type for insulin alarm response
interface InsulinAlarmResponse {
  level: number; // Blood sugar value
  state: string; // State of blood sugar (e.g., "low", "normal", etc.)
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
          process.env.NEXT_PUBLIC_BASE_API_URL + "/insulin-alarm"
        );
        const data: InsulinAlarmResponse = await response.json();

        // Get the current timestamp
        const timestamp = new Date().toISOString();

        // Add the new data point to the graph
        setChartData((prevData) => ({
          labels: [...prevData.labels, timestamp],
          datasets: [
            {
              ...prevData.datasets[0],
              data: [...prevData.datasets[0].data, data.level],
            },
          ],
        }));
      } catch (error) {
        console.error("Error fetching data:", error);
      }
    };

    // Start polling every second
    const intervalId = setInterval(fetchData, 1000);

    // Cleanup interval on component unmount
    return () => clearInterval(intervalId);
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
                unit: "second",
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
