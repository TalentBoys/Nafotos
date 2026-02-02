import { useState, useEffect } from 'react';

interface ResponsiveBreakpoints<T> {
  mobile: T;    // < 768px
  tablet: T;    // 768px - 1024px
  desktop: T;   // > 1024px
}

/**
 * Hook to get a responsive value based on current viewport width
 * Uses matchMedia for efficient breakpoint detection
 */
export function useResponsiveValue<T>(breakpoints: ResponsiveBreakpoints<T>): T {
  const getInitialValue = (): T => {
    if (typeof window === 'undefined') {
      return breakpoints.desktop; // SSR default
    }

    const width = window.innerWidth;
    if (width < 768) return breakpoints.mobile;
    if (width < 1024) return breakpoints.tablet;
    return breakpoints.desktop;
  };

  const [value, setValue] = useState<T>(getInitialValue);

  useEffect(() => {
    // Create media queries
    const mobileQuery = window.matchMedia('(max-width: 767px)');
    const tabletQuery = window.matchMedia('(min-width: 768px) and (max-width: 1023px)');
    const desktopQuery = window.matchMedia('(min-width: 1024px)');

    const updateValue = () => {
      if (mobileQuery.matches) {
        setValue(breakpoints.mobile);
      } else if (tabletQuery.matches) {
        setValue(breakpoints.tablet);
      } else if (desktopQuery.matches) {
        setValue(breakpoints.desktop);
      }
    };

    // Set initial value
    updateValue();

    // Add listeners
    mobileQuery.addEventListener('change', updateValue);
    tabletQuery.addEventListener('change', updateValue);
    desktopQuery.addEventListener('change', updateValue);

    return () => {
      mobileQuery.removeEventListener('change', updateValue);
      tabletQuery.removeEventListener('change', updateValue);
      desktopQuery.removeEventListener('change', updateValue);
    };
  }, [breakpoints.mobile, breakpoints.tablet, breakpoints.desktop]);

  return value;
}
