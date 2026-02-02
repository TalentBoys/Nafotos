import { useMemo } from 'react';
import { File } from '../types';

export interface LayoutPhoto {
  file: File;
  width: number;   // calculated display width
  height: number;  // calculated display height
}

export interface LayoutRow {
  photos: LayoutPhoto[];
  height: number;  // final row height
}

/**
 * Hook to calculate justified layout for photos
 * Similar to Google Photos layout algorithm
 */
export function useJustifiedLayout(
  files: File[],
  containerWidth: number,
  targetHeight: number,
  gap: number = 4
): LayoutRow[] {
  return useMemo(() => {
    if (!files.length || containerWidth === 0) {
      return [];
    }

    const rows: LayoutRow[] = [];
    let currentRow: File[] = [];
    let currentRowWidth = 0;

    // Maximum and minimum width constraints (as percentage of container)
    const MAX_PHOTO_WIDTH = containerWidth * 0.4;
    const MIN_PHOTO_WIDTH = containerWidth * 0.1;

    for (let i = 0; i < files.length; i++) {
      const file = files[i];

      // Get aspect ratio, default to 1:1 if dimensions missing
      let aspectRatio = 1;
      if (file.width && file.height && file.height > 0) {
        aspectRatio = file.width / file.height;
      } else {
        console.warn(`File ${file.filename} missing dimensions, using 1:1 aspect ratio`);
      }

      // Calculate width this photo would take at target height
      const photoWidth = aspectRatio * targetHeight;

      // Check if adding this photo would overflow the row
      const gapWidth = currentRow.length * gap;
      const potentialWidth = currentRowWidth + photoWidth + gapWidth;

      if (currentRow.length > 0 && potentialWidth > containerWidth) {
        // Finalize current row
        rows.push(createJustifiedRow(currentRow, containerWidth, targetHeight, gap, MAX_PHOTO_WIDTH, MIN_PHOTO_WIDTH));
        currentRow = [file];
        currentRowWidth = photoWidth;
      } else {
        // Add to current row
        currentRow.push(file);
        currentRowWidth += photoWidth;
      }
    }

    // Handle last row
    if (currentRow.length > 0) {
      // Don't justify last row if it has fewer than 3 photos
      if (currentRow.length < 3) {
        rows.push(createNonJustifiedRow(currentRow, targetHeight, gap));
      } else {
        rows.push(createJustifiedRow(currentRow, containerWidth, targetHeight, gap, MAX_PHOTO_WIDTH, MIN_PHOTO_WIDTH));
      }
    }

    return rows;
  }, [files, containerWidth, targetHeight, gap]);
}

/**
 * Create a justified row where photos scale to fill container width
 */
function createJustifiedRow(
  rowFiles: File[],
  containerWidth: number,
  targetHeight: number,
  gap: number,
  maxWidth: number,
  minWidth: number
): LayoutRow {
  // Calculate total aspect ratio for the row
  const aspectRatios = rowFiles.map(file => {
    if (file.width && file.height && file.height > 0) {
      return file.width / file.height;
    }
    return 1;
  });

  const sumOfAspectRatios = aspectRatios.reduce((sum, ratio) => sum + ratio, 0);
  const totalGapWidth = (rowFiles.length - 1) * gap;

  // Calculate actual row height to fit container width
  const actualRowHeight = (containerWidth - totalGapWidth) / sumOfAspectRatios;

  // Create layout photos with calculated dimensions
  const photos: LayoutPhoto[] = rowFiles.map((file, index) => {
    let displayWidth = aspectRatios[index] * actualRowHeight;

    // Apply width constraints
    displayWidth = Math.max(minWidth, Math.min(maxWidth, displayWidth));

    return {
      file,
      width: displayWidth,
      height: actualRowHeight
    };
  });

  // Adjust widths if constraints were applied
  const totalCalculatedWidth = photos.reduce((sum, p) => sum + p.width, 0) + totalGapWidth;
  if (Math.abs(totalCalculatedWidth - containerWidth) > 1) {
    // Scale photos proportionally to fit exactly
    const scale = (containerWidth - totalGapWidth) / photos.reduce((sum, p) => sum + p.width, 0);
    photos.forEach(photo => {
      photo.width *= scale;
    });
  }

  return {
    photos,
    height: actualRowHeight
  };
}

/**
 * Create a non-justified row (left-aligned with target height)
 * Used for the last row when it has few photos
 */
function createNonJustifiedRow(
  rowFiles: File[],
  targetHeight: number,
  gap: number
): LayoutRow {
  const photos: LayoutPhoto[] = rowFiles.map(file => {
    let aspectRatio = 1;
    if (file.width && file.height && file.height > 0) {
      aspectRatio = file.width / file.height;
    }

    return {
      file,
      width: aspectRatio * targetHeight,
      height: targetHeight
    };
  });

  return {
    photos,
    height: targetHeight
  };
}
