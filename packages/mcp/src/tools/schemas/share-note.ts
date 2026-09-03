import { z } from "zod";

export const shareNoteInputSchema = {
  text: z.string().min(1).max(2000),
};

export const listProjectNotesInputSchema = {
  limit: z.number().int().positive().max(50).optional(),
};
