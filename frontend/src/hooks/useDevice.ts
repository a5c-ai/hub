'use client';

import { useState, useEffect } from 'react';

export interface DeviceInfo {
  isMobile: boolean;
  isTablet: boolean;
  isDesktop: boolean;
  isTouchDevice: boolean;
  width: number;
  height: number;
}

export function useDevice(): DeviceInfo {
  const [deviceInfo, setDeviceInfo] = useState<DeviceInfo>({
    isMobile: false,
    isTablet: false,
    isDesktop: true,
    isTouchDevice: false,
    width: 1024,
    height: 768,
  });

  useEffect(() => {
    const updateDeviceInfo = () => {
      const width = window.innerWidth;
      const height = window.innerHeight;
      const isTouchDevice = 'ontouchstart' in window || navigator.maxTouchPoints > 0;
      
      // Breakpoints: mobile: < 768px, tablet: 768px - 1023px, desktop: >= 1024px  
      const isMobile = width < 768;
      const isTablet = width >= 768 && width < 1024;
      const isDesktop = width >= 1024;

      setDeviceInfo({
        isMobile,
        isTablet,
        isDesktop,
        isTouchDevice,
        width,
        height,
      });
    };

    // Initial check
    updateDeviceInfo();

    // Listen for resize events
    window.addEventListener('resize', updateDeviceInfo);
    
    return () => {
      window.removeEventListener('resize', updateDeviceInfo);
    };
  }, []);

  return deviceInfo;
}

export function useMobile(): boolean {
  const { isMobile } = useDevice();
  return isMobile;
}

export function useTablet(): boolean {
  const { isTablet } = useDevice();
  return isTablet;
}

export function useDesktop(): boolean {
  const { isDesktop } = useDevice();
  return isDesktop;
}

export function useTouchDevice(): boolean {
  const { isTouchDevice } = useDevice();
  return isTouchDevice;
}