import { Card } from "../../components/ui/Card";
import { StatCard } from "../../components/productivity/StatCard";

export function TaskThroughputReport({ count }: { count: number }) {
  return (
    <Card title="Task throughput" headingLevel={2}>
      <p className="secondary report-caption">Tasks that entered Done within the range.</p>
      <StatCard label="Tasks completed" value={count} sublabel="entered Done in this range" tint="success" />
    </Card>
  );
}
