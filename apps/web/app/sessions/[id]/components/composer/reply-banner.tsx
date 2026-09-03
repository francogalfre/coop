"use client";

import { AnimatePresence, motion } from "motion/react";
import { IconClose, IconReply } from "@/components/icons";

export function ReplyBanner({ seq, onDismiss }: { seq: number | undefined; onDismiss?: () => void }) {
  return (
    <AnimatePresence>
      {seq !== undefined && (
        <motion.div
          initial={{ opacity: 0, y: 4 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: 4 }}
          transition={{ duration: 0.2, ease: [0.16, 1, 0.3, 1] }}
          className="mb-1.5 flex items-center gap-1.5 text-xs text-human"
        >
          <IconReply size={12} />
          <span>replying to step {seq}</span>
          <button
            type="button"
            onClick={onDismiss}
            className="ml-0.5 grid size-4 place-items-center rounded text-human/70 hover:bg-human/15 hover:text-human"
          >
            <IconClose size={10} />
          </button>
        </motion.div>
      )}
    </AnimatePresence>
  );
}
