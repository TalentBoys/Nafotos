import { useState, useEffect, useRef, RefObject } from 'react';

/**
 * Hook to track container width using ResizeObserver
 * Returns the current width of the referenced container
 */
export function useContainerWidth<T extends HTMLElement>(): [RefObject<T>, number] {
  const containerRef = useRef<T>(null);
  const [width, setWidth] = useState<number>(0);
  const debounceTimeoutRef = useRef<number | null>(null);

  useEffect(() => {
    const element = containerRef.current;
    if (!element) return;

    // Set initial width
    setWidth(element.offsetWidth);

    // Create ResizeObserver with debouncing
    const resizeObserver = new ResizeObserver((entries) => {
      if (debounceTimeoutRef.current) {
        clearTimeout(debounceTimeoutRef.current);
      }

      debounceTimeoutRef.current = setTimeout(() => {
        for (const entry of entries) {
          const newWidth = entry.contentRect.width;
          setWidth(newWidth);
        }
      }, 300); // 300ms debounce
    });

    resizeObserver.observe(element);

    return () => {
      if (debounceTimeoutRef.current) {
        clearTimeout(debounceTimeoutRef.current);
      }
      resizeObserver.disconnect();
    };
  }, []);

  return [containerRef, width];
}
