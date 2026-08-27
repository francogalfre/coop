import { DocPageHeader } from "../components/doc-page-header";
import { CliCommand } from "../components/cli-command";

export default function DocsCliReference() {
  return (
    <>
      <DocPageHeader
        eyebrow="CLI Reference"
        title="coop commands"
        intro="Four commands, verified against packages/cli/cmd/coop. Nothing here is invented — if a flag isn't listed, it doesn't exist yet."
      />

      <div className="space-y-6">
        <CliCommand
          name="coop login"
          description="Authenticates with GitHub's device flow and saves a coop credential to disk. Required before --project can be used with attach or run."
          usage="coop login"
          example="coop login"
        />

        <CliCommand
          name="coop attach"
          description="Keeps your normal agent TUI running as-is. Detects the harness in the current directory, installs its native hook or plugin, and starts a local server that receives, redacts, and forwards events to the relay."
          usage="coop attach [--harness=<name>] --project=<slug>"
          flags={[
            {
              flag: "--harness=<name>",
              description:
                "Install hooks for only this harness (claude-code, opencode, pi). Required if more than one harness is detected in the working directory.",
            },
            {
              flag: "--project=<slug>",
              description:
                "Required — every session must belong to a coop project. Implies you've already run coop login.",
            },
          ]}
          example="coop attach --project=my-app"
          note="Prints a share link (http://localhost:3000/sessions/<id>?token=<token>) and keeps running until you press Ctrl-C."
        />

        <CliCommand
          name="coop run"
          description="Wraps <cmd> in a pty coop owns. Same hook wiring as attach, but because coop also owns stdin, steering lands immediately — even for harnesses with no native injection primitive."
          usage="coop run [--harness=<name>] --project=<slug> -- <cmd> [args...]"
          flags={[
            {
              flag: "--harness=<name>",
              description: "Install hooks for only this harness (claude-code, opencode, pi).",
            },
            {
              flag: "--project=<slug>",
              description:
                "Required — every session must belong to a coop project. Implies you've already run coop login.",
            },
          ]}
          example="coop run --project=my-app -- opencode"
          note="Omit -- <cmd> entirely and coop run launches claude by default."
        />

        <CliCommand
          name="coop detach"
          description="Removes any coop hook or plugin entries left behind in a directory. Use it after coop attach exits uncleanly (crash, kill -9) to clean up."
          usage="coop detach [dir]"
          flags={[{ flag: "[dir]", description: "Directory to clean up. Defaults to the current directory." }]}
          example="coop detach"
        />
      </div>
    </>
  );
}
