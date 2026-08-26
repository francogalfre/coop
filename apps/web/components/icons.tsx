import { HugeiconsIcon, type HugeiconsIconProps } from "@hugeicons/react";
import {
  Add01Icon,
  Alert02Icon,
  ArrowLeft01Icon,
  ArrowRight01Icon,
  ArrowUpRight01Icon,
  Cancel01Icon,
  CheckmarkCircle02Icon,
  Clock01Icon,
  Copy01Icon,
  File01Icon,
  Folder01Icon,
  Github01Icon,
  Link01Icon,
  Loading03Icon,
  Logout01Icon,
  Message01Icon,
  PencilEdit02Icon,
  Robot01Icon,
  Sent02Icon,
  SourceCodeIcon,
  SparklesIcon,
  TerminalIcon,
  Tick02Icon,
  UserGroupIcon,
} from "@hugeicons/core-free-icons";

type IconProps = Omit<HugeiconsIconProps, "icon">;

function makeIcon(icon: HugeiconsIconProps["icon"]) {
  return function Icon(props: IconProps) {
    return <HugeiconsIcon icon={icon} strokeWidth={1.8} {...props} />;
  };
}

export const IconClose = makeIcon(Cancel01Icon);
export const IconCheck = makeIcon(Tick02Icon);
export const IconCheckCircle = makeIcon(CheckmarkCircle02Icon);
export const IconChevronRight = makeIcon(ArrowRight01Icon);
export const IconChevronLeft = makeIcon(ArrowLeft01Icon);
export const IconExternal = makeIcon(ArrowUpRight01Icon);
export const IconPlus = makeIcon(Add01Icon);
export const IconCopy = makeIcon(Copy01Icon);
export const IconLink = makeIcon(Link01Icon);
export const IconGithub = makeIcon(Github01Icon);
export const IconLogout = makeIcon(Logout01Icon);
export const IconMessage = makeIcon(Message01Icon);
export const IconSend = makeIcon(Sent02Icon);
export const IconAgent = makeIcon(Robot01Icon);
export const IconTerminal = makeIcon(TerminalIcon);
export const IconCode = makeIcon(SourceCodeIcon);
export const IconFile = makeIcon(File01Icon);
export const IconFolder = makeIcon(Folder01Icon);
export const IconEdit = makeIcon(PencilEdit02Icon);
export const IconPeople = makeIcon(UserGroupIcon);
export const IconClock = makeIcon(Clock01Icon);
export const IconSpinner = makeIcon(Loading03Icon);
export const IconSparkles = makeIcon(SparklesIcon);
export const IconAlert = makeIcon(Alert02Icon);
