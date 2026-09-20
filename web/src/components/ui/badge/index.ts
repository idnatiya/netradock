import { cva, type VariantProps } from 'class-variance-authority'

export const badgeVariants = cva(
  'inline-flex items-center gap-1.5 rounded-md border px-2 py-0.5 text-xs font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2',
  {
    variants: {
      variant: {
        default: 'border-transparent bg-primary text-primary-foreground shadow hover:bg-primary/80',
        secondary: 'border-transparent bg-secondary text-secondary-foreground hover:bg-secondary/80',
        destructive: 'border-destructive/25 bg-destructive/15 text-destructive dark:text-rose-400',
        outline: 'border-border text-foreground',
        success: 'border-emerald-500/25 bg-emerald-500/15 text-emerald-600 dark:text-emerald-400',
        warning: 'border-amber-500/25 bg-amber-500/15 text-amber-600 dark:text-amber-400',
        info: 'border-sky-500/25 bg-sky-500/15 text-sky-600 dark:text-sky-400',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  },
)

export type BadgeVariants = VariantProps<typeof badgeVariants>
export { default as Badge } from './Badge.vue'
