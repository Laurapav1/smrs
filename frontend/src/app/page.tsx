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
  age: number; // Age of the person
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

  const [currentState, setCurrentState] = useState<InsulinAlarmResponse>({
    level: 0,
    state: "unknown",
    age: 30, // Default age
  });

  const [selectedAge, setSelectedAge] = useState<number>(30); // Default age

  // Helper function to get color based on state
  const getStateColor = (state: string): string => {
    switch (state) {
      case "low":
        return "blue";
      case "normal":
        return "green";
      case "high":
        return "orange";
      case "critical":
        return "red";
      default:
        return "gray";
    }
  };

  useEffect(() => {
    const fetchData = async () => {
      try {
        const response = await fetch(
          `${process.env.NEXT_PUBLIC_BASE_API_URL}/insulin-alarm?age=${selectedAge}`
        );
        const data: InsulinAlarmResponse = await response.json();

        // Get the current timestamp
        const timestamp = new Date().toISOString();

        // Add the new data point and limit to 30 data points
        setChartData((prevData) => {
          const newLabels = [...prevData.labels, timestamp];
          const newData = [...prevData.datasets[0].data, data.level];

          // Limit the arrays to 30 entries
          return {
            labels: newLabels.slice(-30),
            datasets: [
              {
                ...prevData.datasets[0],
                data: newData.slice(-30),
              },
            ],
          };
        });

        // Update the current state
        setCurrentState(data);
      } catch (error) {
        console.error("Error fetching data:", error);
      }
    };

    // Start polling every second
    const intervalId = setInterval(fetchData, 1000);

    // Cleanup interval on component unmount
    return () => clearInterval(intervalId);
  }, [selectedAge]); // Re-fetch data if the selected age changes

  return (
    <div>
      <h2>Blood Sugar Levels</h2>
      <div style={{ marginBottom: "1rem" }}>
        <label htmlFor="age-range" style={{ marginRight: "0.5rem" }}>
          Select Your Age Range:
        </label>
        <select
          id="age-range"
          value={selectedAge}
          onChange={(e) => setSelectedAge(Number(e.target.value))}
          style={{
            padding: "0.5rem",
            fontSize: "1rem",
            borderRadius: "4px",
            border: "1px solid #ccc",
          }}
        >
          <option value={10}>Under 18</option>
          <option value={30}>18 - 65</option>
          <option value={70}>Over 65</option>
        </select>
      </div>
      <p
        style={{
          fontWeight: "bold",
          fontSize: "1.5rem",
          color: getStateColor(currentState.state), // Apply color based on state
        }}
      >
        Current State: {currentState.state.toUpperCase()} (
        {currentState.level.toFixed(1)}) for Age {currentState.age}
      </p>
      <Line
        data={chartData}
        options={{
          datasets: {
            line: {
              pointHitRadius: 50,
            },
          },
          hover: {},
          animation: false,
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
