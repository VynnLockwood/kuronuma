import { notFound } from "next/navigation";

// Any unknown address under /th lands here so the Thai not-found page (app/th/not-found.tsx) is the one rendered.
export default function MissingTh() {
  notFound();
}
