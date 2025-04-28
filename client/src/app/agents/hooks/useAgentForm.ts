import { useState } from "react";
import type { Agent } from "./useAgentsManager";

export function useAgentForm() {
  const [formData, setFormData] = useState({
    name: "",
    role: "validator",
    style: "",
    mood: "",
  });
  const [traits, setTraits] = useState<string[]>([]);
  const [influences, setInfluences] = useState<string[]>([]);
  const [newTrait, setNewTrait] = useState("");
  const [newInfluence, setNewInfluence] = useState("");

  // Updates form data when input fields change
  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>
  ) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  // Adds a new trait to the traits array
  const addTrait = () => {
    if (newTrait.trim() !== "") {
      setTraits([...traits, newTrait.trim()]);
      setNewTrait("");
    }
  };

  // Removes a trait at the specified index
  const removeTrait = (index: number) => {
    setTraits((prev) => prev.filter((_, i) => i !== index));
  };

  // Adds a new influence to the influences array
  const addInfluence = () => {
    if (newInfluence.trim() !== "") {
      setInfluences([...influences, newInfluence.trim()]);
      setNewInfluence("");
    }
  };

  // Removes an influence at the specified index
  const removeInfluence = (index: number) => {
    setInfluences((prev) => prev.filter((_, i) => i !== index));
  };

  // Resets all form fields to their initial state
  const resetForm = () => {
    setFormData({
      name: "",
      role: "producer",
      style: "",
      mood: "",
    });
    setTraits([]);
    setInfluences([]);
    setNewTrait("");
    setNewInfluence("");
  };

  // Creates a new agent object from the form data
  const buildAgent = (): Agent => ({
    id: Date.now().toString(),
    name: formData.name,
    role: formData.role as "producer" | "validator",
    traits,
    style: formData.style,
    influences,
    mood: formData.mood,
  });

  return {
    formData,
    setFormData,
    traits,
    influences,
    newTrait,
    newInfluence,
    setNewTrait,
    setNewInfluence,
    handleChange,
    addTrait,
    removeTrait,
    addInfluence,
    removeInfluence,
    resetForm,
    buildAgent,
  };
}