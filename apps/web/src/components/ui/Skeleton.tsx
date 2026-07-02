export function SkeletonBlock({ lines = 3 }: { lines?: number }) {
  return (
    <div className="skeletonBlock" aria-hidden="true">
      {Array.from({ length: Math.max(1, lines) }).map((_, index) => (
        <span className="skeletonLine" key={index} />
      ))}
    </div>
  );
}

export default SkeletonBlock;
