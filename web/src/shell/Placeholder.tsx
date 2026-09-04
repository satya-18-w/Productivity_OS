import { PageHeader } from "../components/layout/PageHeader";
import { Card } from "../components/ui/Card";
import { ScreenLayout } from "./ScreenLayout";

export interface PlaceholderProps {
  /** Screen name, e.g. "Goals". */
  name: string;
  /** Which implementation phase builds it. */
  phase: number;
}

/**
 * Temporary content for a route whose real screen is not built yet. Confirms
 * the shell + routing work end to end.
 */
export function Placeholder({ name, phase }: PlaceholderProps) {
  return (
    <ScreenLayout>
      <PageHeader title={name} subtitle={`This screen is built in phase ${phase}.`} />
      <Card>
        <p className="secondary">
          Routing and the app shell are in place. The <strong>{name}</strong> screen lands
          in phase {phase} of the frontend implementation plan.
        </p>
      </Card>
    </ScreenLayout>
  );
}
