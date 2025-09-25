import { h } from "preact";
import { useEffect, useRef } from "preact/hooks";
import * as d3 from "d3";
import { FlameGraphNode } from "../lib/api";

interface FlameGraphProps {
  data: FlameGraphNode | null;
  width?: number;
  height?: number;
}

export function FlameGraph({
  data,
  width = 900,
  height = 400,
}: FlameGraphProps) {
  const svgRef = useRef<SVGSVGElement>(null);
  const tooltipRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!data || !svgRef.current) return;

    // Clear previous content
    const svg = d3.select(svgRef.current);
    svg.selectAll("*").remove();

    // Set up dimensions
    const cellHeight = 20;
    const actualHeight = height || 400;

    // Create hierarchy
    const root = d3
      .hierarchy(data)
      .sum((d: any) => d.value || 0)
      .sort((a, b) => (b.value || 0) - (a.value || 0));

    // Create partition layout
    const partition = d3
      .partition<FlameGraphNode>()
      .size([width, actualHeight])
      .padding(1)
      .round(true);

    partition(root);

    // Color scale
    const color = d3.scaleOrdinal(d3.schemeTableau10);

    // Create groups for each node
    const g = svg
      .selectAll("g")
      .data(root.descendants())
      .join("g")
      .attr("transform", (d) => `translate(${d.x0},${d.depth * cellHeight})`);

    // Add rectangles
    g.append("rect")
      .attr("x", 0)
      .attr("width", (d) => Math.max(0, d.x1 - d.x0))
      .attr("height", cellHeight - 1)
      .attr("fill", (d) => {
        if (!d.depth) return "#f3f4f6";
        return color(d.data.name);
      })
      .style("stroke", "#fff")
      .style("cursor", "pointer")
      .on("mouseover", function (event, d) {
        if (tooltipRef.current) {
          const percentage = (
            ((d.value || 0) / (root.value || 1)) *
            100
          ).toFixed(2);
          tooltipRef.current.innerHTML = `
            <div style="font-weight: bold;">${d.data.name}</div>
            <div>${percentage}% of total</div>
            <div>Value: ${d.value}</div>
          `;
          tooltipRef.current.style.display = "block";
          tooltipRef.current.style.left = event.pageX + 10 + "px";
          tooltipRef.current.style.top = event.pageY - 28 + "px";
        }
      })
      .on("mousemove", function (event) {
        if (tooltipRef.current) {
          tooltipRef.current.style.left = event.pageX + 10 + "px";
          tooltipRef.current.style.top = event.pageY - 28 + "px";
        }
      })
      .on("mouseout", function () {
        if (tooltipRef.current) {
          tooltipRef.current.style.display = "none";
        }
      });

    // Add text labels
    g.append("text")
      .attr("x", 4)
      .attr("y", cellHeight / 2)
      .attr("dy", "0.32em")
      .text((d) => {
        const width = d.x1 - d.x0;
        if (width < 30) return "";
        const name = d.data.name;
        const maxChars = Math.floor(width / 7);
        return name.length > maxChars
          ? name.substring(0, maxChars - 1) + "…"
          : name;
      })
      .style("pointer-events", "none")
      .style("fill", (d) => (!d.depth ? "#000" : "#fff"))
      .style("font-size", "12px")
      .style("font-family", "monospace");
  }, [data, width, height]);

  if (!data) {
    return (
      <div className="flex items-center justify-center h-64 text-muted-foreground">
        No flame graph data available
      </div>
    );
  }

  return (
    <div className="relative">
      <svg
        ref={svgRef}
        width={width}
        height={height}
        style={{ width: "100%", height: "auto" }}
        viewBox={`0 0 ${width} ${height}`}
      />
      <div
        ref={tooltipRef}
        className="absolute bg-gray-900 text-white p-2 rounded shadow-lg text-sm"
        style={{
          display: "none",
          pointerEvents: "none",
          zIndex: 1000,
          position: "fixed",
        }}
      />
    </div>
  );
}
