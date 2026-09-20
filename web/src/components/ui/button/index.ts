import { type HTMLAttributes } from 'vue'
import { cva, type VariantProps } from 'class-variance-authority'

export const buttonVariants = cva(
  'inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0 cursor-pointer select-none',
  {
    variants: {
      variant: {
        default: 'bg-primary text-primary-foreground shadow-sm hover:bg-primary/90 border border-primary/20 active:scale-[0.98]',
        destructive: 'bg-destructive text-destructive-foreground shadow-sm hover:bg-destructive/90 active:scale-[0.98]',
        outline: 'border border-border bg-card shadow-2xs hover:bg-accent hover:text-accent-foreground text-foreground active:scale-[0.98]',
        secondary: 'bg-secondary text-secondary-foreground shadow-2xs hover:bg-secondary/80 text-foreground active:scale-[0.98]',
        ghost: 'hover:bg-accent hover:text-accent-foreground active:scale-[0.98]',
        link: 'text-primary underline-offset-4 hover:underline',
        success: 'bg-emerald-600 text-white shadow-sm hover:bg-emerald-700 active:scale-[0.98]',
      },
      size: {
        default: 'h-9.5 px-4 py-2 text-sm',
        sm: 'h-8.5 rounded-md px-3 text-xs',
        lg: 'h-10 rounded-md px-8 text-base',
        icon: 'h-9 w-9',
        'icon-sm': 'h-7.5 w-7.5',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  },
)

export type ButtonVariants = VariantProps<typeof buttonVariants>

export { default as Button } from './Button.vue'
