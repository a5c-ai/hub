import { Fragment } from 'react';
import { Menu, Transition } from '@headlessui/react';
import { ChevronDownIcon } from '@heroicons/react/20/solid';
import { cn } from '@/lib/utils';

interface DropdownItem {
  label: string;
  onClick: () => void;
  icon?: React.ReactNode;
  disabled?: boolean;
  destructive?: boolean;
}

interface DropdownProps {
  trigger: React.ReactNode;
  items: DropdownItem[];
  align?: 'left' | 'right';
  className?: string;
}

export function Dropdown({ trigger, items, align = 'right', className }: DropdownProps) {
  return (
    <Menu as="div" className={cn('relative inline-block text-left', className)}>
      <Menu.Button as="div">
        {trigger}
      </Menu.Button>

      <Transition
        as={Fragment}
        enter="transition ease-out duration-100"
        enterFrom="transform opacity-0 scale-95"
        enterTo="transform opacity-100 scale-100"
        leave="transition ease-in duration-75"
        leaveFrom="transform opacity-100 scale-100"
        leaveTo="transform opacity-0 scale-95"
      >
        <Menu.Items
          className={cn(
            'absolute z-50 mt-2 w-56 origin-top-right divide-y divide-border rounded-md bg-popover shadow-lg ring-1 ring-border ring-opacity-5 focus:outline-none',
            align === 'left' ? 'left-0 origin-top-left' : 'right-0 origin-top-right'
          )}
        >
          <div className="px-1 py-1">
            {items.map((item, index) => (
              <Menu.Item key={index} disabled={item.disabled}>
                {({ active }) => (
                  <button
                    className={cn(
                      'group flex w-full items-center rounded-md px-2 py-2 text-sm transition-colors',
                      active && !item.disabled
                        ? 'bg-muted text-foreground'
                        : 'text-muted-foreground',
                      item.disabled && 'opacity-50 cursor-not-allowed',
                      item.destructive && active && 'bg-destructive text-destructive-foreground',
                      item.destructive && !active && 'text-destructive'
                    )}
                    onClick={item.onClick}
                    disabled={item.disabled}
                  >
                    {item.icon && (
                      <div className="mr-2 h-4 w-4" aria-hidden="true">
                        {item.icon}
                      </div>
                    )}
                    {item.label}
                  </button>
                )}
              </Menu.Item>
            ))}
          </div>
        </Menu.Items>
      </Transition>
    </Menu>
  );
}

// Simple dropdown button variant
interface DropdownButtonProps extends Omit<DropdownProps, 'trigger'> {
  label: string;
  variant?: 'default' | 'outline' | 'ghost';
  size?: 'sm' | 'default' | 'lg';
}

export function DropdownButton({
  label,
  items,
  variant = 'default',
  size = 'default',
  ...props
}: DropdownButtonProps) {
  const buttonClasses = cn(
    'inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50',
    {
      'bg-primary text-primary-foreground hover:bg-primary/90': variant === 'default',
      'border border-input bg-background hover:bg-muted hover:text-muted-foreground': variant === 'outline',
      'hover:bg-muted hover:text-muted-foreground': variant === 'ghost',
    },
    {
      'h-9 px-3': size === 'sm',
      'h-10 px-4 py-2': size === 'default',
      'h-11 px-8': size === 'lg',
    }
  );

  return (
    <Dropdown
      trigger={
        <button className={buttonClasses}>
          {label}
          <ChevronDownIcon className="ml-2 h-4 w-4" aria-hidden="true" />
        </button>
      }
      items={items}
      {...props}
    />
  );
}