import { useEffect, useState } from "react";

import defaultServiceImage from "../../assets/default-service.svg";

interface ServiceImageProps {
  imageUrl: string;
  alt: string;
  className?: string;
}

export const ServiceImage = ({ imageUrl, alt, className = "service-thumb" }: ServiceImageProps): JSX.Element => {
  const resolvedSource = imageUrl.trim().length > 0 ? imageUrl : defaultServiceImage;
  const [source, setSource] = useState(resolvedSource);

  useEffect(() => {
    setSource(resolvedSource);
  }, [resolvedSource]);

  return (
    <img
      className={className}
      src={source}
      alt={alt}
      onError={() => {
        if (source !== defaultServiceImage) {
          setSource(defaultServiceImage);
        }
      }}
    />
  );
};
